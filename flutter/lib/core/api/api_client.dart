import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app.dart';
import '../auth/auth_storage.dart';
import '../routes/app_routes.dart';
import 'api_exception.dart';

/// Endpoint base del backend Stasera.
/// In produzione va sostituito con l'URL del deploy Railway.
const String kApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8080',
);

/// Provider singleton per il client Dio configurato con interceptor JWT
/// e refresh automatico thread-safe su 401.
final apiClientProvider = Provider<Dio>((ref) {
  final storage = ref.watch(authStorageProvider);

  final dio = Dio(_baseOptions());
  dio.interceptors.add(AuthInterceptor(storage: storage, ref: ref));
  return dio;
});

/// Opzioni base condivise tra il Dio principale e il Dio "refresh" (che non
/// monta l'interceptor per evitare loop). Centralizzare qui impedisce drift di
/// timeout/baseUrl/headers tra le due istanze.
BaseOptions _baseOptions() => BaseOptions(
      baseUrl: '$kApiBaseUrl/api/v1',
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Content-Type': 'application/json'},
    );

/// Una richiesta in attesa di essere ritentata dopo un refresh del token.
class _PendingRequest {
  _PendingRequest(this.options, this.handler);
  final RequestOptions options;
  final ErrorInterceptorHandler handler;
}

/// Interceptor che:
/// 1. Aggiunge `Authorization: Bearer <access_token>` a ogni richiesta.
/// 2. Su risposta 401, serializza i refresh: il primo 401 innesca UN refresh
///    tramite l'endpoint `/auth/refresh` (con un Dio separato, senza interceptor,
///    per evitare loop); i 401 concorrenti vengono accodati e, a refresh
///    avvenuto, ritentati con il nuovo access token.
/// 3. Se il refresh fallisce, svuota i token e naviga a `/login` via GoRouter.
class AuthInterceptor extends Interceptor {
  AuthInterceptor({required this.storage, required this.ref});

  final AuthStorage storage;
  final Ref ref;

  /// Non-null mentre un refresh è in corso; i 401 concorrenti lo attendono.
  Completer<void>? _refreshCompleter;

  /// Richieste accodate durante un refresh in corso.
  final List<_PendingRequest> _queue = [];

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // Non attaccare il token sulle route di auth pubbliche.
    final isAuthEndpoint = options.path.startsWith('/auth/login') ||
        options.path.startsWith('/auth/register') ||
        options.path.startsWith('/auth/refresh');

    if (!isAuthEndpoint) {
      final token = await storage.getAccessToken();
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final response = err.response;
    if (response == null || response.statusCode != 401) {
      return handler.next(err);
    }

    // Evita loop infiniti sul refresh stesso: se il refresh endpoint ritorna
    // 401 il refresh token è invalido → logout.
    if (response.requestOptions.path.startsWith('/auth/refresh')) {
      await _logoutAndRedirect(handler, err);
      return;
    }

    // Se un refresh è già in corso, accoda questa richiesta e attendi.
    if (_refreshCompleter != null) {
      _queue.add(_PendingRequest(err.requestOptions, handler));
      return;
    }

    // Primo 401: avvia il refresh e guida il replay della coda.
    final completer = Completer<void>();
    _refreshCompleter = completer;
    try {
      final refreshed = await _tryRefresh();
      if (!refreshed) {
        await _logoutAndRedirect(handler, err);
        _drainQueueOnError(err);
        completer.complete();
        return;
      }
      // Replay del richiedente originale.
      await _replayOne(err.requestOptions, handler);
      // Replay della coda accumulata durante il refresh.
      _drainQueueOnSuccess();
      completer.complete();
    } finally {
      _refreshCompleter = null;
    }
  }

  /// Ritenta una singola richiesta con il nuovo access token.
  Future<void> _replayOne(
    RequestOptions options,
    ErrorInterceptorHandler handler,
  ) async {
    try {
      final cloned = await _retry(options);
      handler.resolve(cloned);
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  /// Svuota la coda dopo un refresh riuscito: ogni richiesta in attesa viene
  /// ritentata con il nuovo access token.
  void _drainQueueOnSuccess() {
    if (_queue.isEmpty) return;
    final pending = List<_PendingRequest>.of(_queue);
    _queue.clear();
    for (final p in pending) {
      // Fire-and-forget: ogni replay è indipendente.
      _replayOne(p.options, p.handler);
    }
  }

  /// Svuota la coda propagando l'errore originale a ogni richiesta in attesa.
  void _drainQueueOnError(DioException err) {
    if (_queue.isEmpty) return;
    final pending = List<_PendingRequest>.of(_queue);
    _queue.clear();
    for (final p in pending) {
      p.handler.next(err);
    }
  }

  /// Chiama `/auth/refresh` con un Dio separato (senza interceptor) e persiste
  /// la nuova coppia di token. Ritorna `false` se il refresh non è possibile.
  Future<bool> _tryRefresh() async {
    try {
      final refreshToken = await storage.getRefreshToken();
      if (refreshToken == null) return false;

      final dio = Dio(_baseOptions());
      final resp = await dio.post(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      if (resp.statusCode != 200) return false;

      final data = resp.data as Map<String, dynamic>;
      await storage.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<Response<dynamic>> _retry(RequestOptions options) async {
    final token = await storage.getAccessToken();
    options.headers['Authorization'] = 'Bearer $token';
    // Dio senza interceptor: evita ricorsione su un eventuale nuovo 401.
    return Dio(_baseOptions()).fetch(options);
  }

  /// Svuota i token e naviga a `/login` via GoRouter globale.
  Future<void> _logoutAndRedirect(
    ErrorInterceptorHandler handler,
    DioException err,
  ) async {
    await storage.clear();
    ref.read(routerProvider).go(AppRoutes.login);
    handler.next(err);
  }
}

/// Converte un [DioException] in una [ApiException] leggibile dal UI.
///
/// Funzione unica di conversione errori: i repository la usano per trasformare
/// i `DioException` in `ApiException` con il messaggio safe proveniente dal
/// backend (`{"error":"..."}`).
ApiException toApiException(DioException err) {
  final response = err.response;
  if (response != null) {
    final data = response.data;
    String message = err.message ?? 'Errore di rete';
    if (data is Map<String, dynamic> && data['error'] != null) {
      message = data['error'] as String;
    }
    return ApiException(
      statusCode: response.statusCode ?? 0,
      message: message,
      error: err,
    );
  }
  return ApiException(
    statusCode: 0,
    message: err.message ?? 'Errore di rete',
    error: err,
  );
}
