import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app.dart';
import '../../../core/api/api_exception.dart';
import '../../../core/models/dto.dart';
import '../../../core/notifications/notification_service.dart';
import '../../../core/routes/app_routes.dart';
import '../data/auth_repository.dart';

/// Stato auth esposto al UI.
sealed class AuthState {
  const AuthState();
}

class AuthInitial extends AuthState {
  const AuthInitial();
}

class AuthLoading extends AuthState {
  const AuthLoading();
}

class Authenticated extends AuthState {
  const Authenticated(this.user);
  final UserDto user;
}

class AuthError extends AuthState {
  const AuthError(this.message);
  final String message;
}

class Unauthenticated extends AuthState {
  const Unauthenticated();
}

/// Notifier che orchestra login/register/logout.
///
/// Possiede un [Ref] per poter innescare la navigazione globale al logout
/// (stesso wiring del `AuthInterceptor`: `ref.read(routerProvider).go(...)`),
/// in modo che il logout da settings e il logout da 401-refresh-fallito
/// condividano un unico percorso di navigazione.
class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._repo, this._ref) : super(const AuthInitial());

  final AuthRepository _repo;
  final Ref _ref;

  Future<void> login(String email, String password) async {
    state = const AuthLoading();
    try {
      final auth = await _repo.login(email: email, password: password);
      state = Authenticated(auth.user);
    } on ApiException catch (e) {
      state = AuthError(e.message);
    } catch (_) {
      state = const AuthError('Errore inatteso');
    }
  }

  Future<void> register(String email, String password, String name) async {
    state = const AuthLoading();
    try {
      final auth = await _repo.register(
        email: email,
        password: password,
        displayName: name,
      );
      state = Authenticated(auth.user);
    } on ApiException catch (e) {
      state = AuthError(e.message);
    } catch (_) {
      state = const AuthError('Errore inatteso');
    }
  }

  /// Aggiorna il nome visualizzato dell'utente autenticato.
  /// Restituisce `null` in caso di successo, altrimenti il messaggio d'errore.
  Future<String?> updateDisplayName(String name) async {
    try {
      final updated = await _repo.updateMe(displayName: name);
      state = Authenticated(updated);
      return null;
    } on ApiException catch (e) {
      return e.message;
    } catch (_) {
      return 'Errore inatteso';
    }
  }

  /// Cambia la password dell'utente autenticato.
  /// Restituisce `null` in caso di successo, altrimenti il messaggio d'errore.
  Future<String?> changePassword(String current, String next) async {
    try {
      await _repo.changePassword(currentPassword: current, newPassword: next);
      return null;
    } on ApiException catch (e) {
      if (e.statusCode == 401) {
        return 'Password corrente errata';
      }
      return e.message;
    } catch (_) {
      return 'Errore inatteso';
    }
  }

  /// Logout: cancella le notifiche, i token e naviga a `/login` via GoRouter
  /// globale (wiring condiviso con `AuthInterceptor`).
  Future<void> logout() async {
    await NotificationService.cancelAllReminders();
    await _repo.logout();
    state = const Unauthenticated();
    _ref.read(routerProvider).go(AppRoutes.login);
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.watch(authRepositoryProvider), ref);
});
