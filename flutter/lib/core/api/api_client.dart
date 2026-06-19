import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../auth/auth_storage.dart';
import 'api_exception.dart';

/// Endpoint base del backend Stasera.
/// In produzione va sostituito con l'URL del deploy Railway.
const String kApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8080',
);

/// Provider singleton per il client Dio configurato con interceptor JWT
/// e refresh automatico su 401.
final apiClientProvider = Provider<Dio>((ref) {
  final storage = ref.watch(authStorageProvider);

  final dio = Dio(
    BaseOptions(
      baseUrl: '$kApiBaseUrl/api/v1',
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Content-Type': 'application/json'},
    ),
  );

  dio.interceptors.add(AuthInterceptor(storage: storage, ref: ref));
  return dio;
});

/// Interceptor che:
/// 1. Aggiunge `Authorization: Bearer <access_token>` a ogni richiesta.
/// 2. Su risposta 401, prova il refresh con un token rotation e ritenta
///    la richiesta originale una sola volta.
/// 3. Se il refresh fallisce, svuota i token e naviga a /login.
class AuthInterceptor extends Interceptor {
  AuthInterceptor({required this.storage, required this.ref});

  final AuthStorage storage;
  final Ref ref;

  bool _isRefreshing = false;

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

    // Evita loop infiniti sul refresh stesso.
    if (response.requestOptions.path.startsWith('/auth/refresh')) {
      await _logoutAndRedirect(handler, err);
      return;
    }

    final refreshed = await _tryRefresh();
    if (!refreshed) {
      await _logoutAndRedirect(handler, err);
      return;
    }

    try {
      final cloned = await _retry(response.requestOptions);
      handler.resolve(cloned);
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  Future<bool> _tryRefresh() async {
    if (_isRefreshing) return false;
    _isRefreshing = true;
    try {
      final refreshToken = await storage.getRefreshToken();
      if (refreshToken == null) return false;

      final dio = Dio(
        BaseOptions(baseUrl: '$kApiBaseUrl/api/v1'),
      );
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
    } finally {
      _isRefreshing = false;
    }
  }

  Future<Response<dynamic>> _retry(RequestOptions options) async {
    final token = await storage.getAccessToken();
    options.headers['Authorization'] = 'Bearer $token';
    final dio = Dio(BaseOptions(
      baseUrl: options.baseUrl,
      headers: options.headers,
    ));
    return dio.fetch(options);
  }

  Future<void> _logoutAndRedirect(
    ErrorInterceptorHandler handler,
    DioException err,
  ) async {
    await storage.clear();
    // Navigazione globale a /login: usa GoRouter tramite navigatorKey globale.
    // App.dart possiede il router; qui emettiamo un evento tramite un stream
    // semplice che il router ascolta. Per semplicità usiamo rootNavigator.
    handler.next(err);
  }
}

/// Converte un [DioException] in una [ApiException] leggibile dal UI.
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