import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_exception.dart';
import '../../../core/models/dto.dart';
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
class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._repo) : super(const AuthInitial());

  final AuthRepository _repo;

  Future<void> checkSession() async {
    state = const AuthLoading();
    try {
      final user = await _repo.me();
      state = Authenticated(user);
    } catch (_) {
      state = const Unauthenticated();
    }
  }

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

  Future<void> logout() async {
    await _repo.logout();
    state = const Unauthenticated();
  }
}

final authProvider =
    StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.watch(authRepositoryProvider));
});