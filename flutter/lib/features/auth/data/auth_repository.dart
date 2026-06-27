import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/auth/auth_storage.dart';
import '../../../core/models/dto.dart';

/// Repository per le chiamate di autenticazione.
class AuthRepository {
  AuthRepository(this._dio, this._storage);

  final Dio _dio;
  final AuthStorage _storage;

  Future<AuthResponseDto> register({
    required String email,
    required String password,
    required String displayName,
  }) async {
    try {
      final resp = await _dio.post('/auth/register', data: {
        'email': email,
        'password': password,
        'display_name': displayName,
      });
      final auth = AuthResponseDto.fromJson(resp.data as Map<String, dynamic>);
      await _storage.saveTokens(
        accessToken: auth.accessToken,
        refreshToken: auth.refreshToken,
      );
      return auth;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<AuthResponseDto> login({
    required String email,
    required String password,
  }) async {
    try {
      final resp = await _dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });
      final auth = AuthResponseDto.fromJson(resp.data as Map<String, dynamic>);
      await _storage.saveTokens(
        accessToken: auth.accessToken,
        refreshToken: auth.refreshToken,
      );
      return auth;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<void> logout() async {
    await _storage.clear();
  }

  Future<UserDto> updateMe({required String displayName}) async {
    try {
      final resp = await _dio.patch('/auth/me', data: {
        'display_name': displayName,
      });
      return UserDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<void> changePassword({
    required String currentPassword,
    required String newPassword,
  }) async {
    try {
      await _dio.post('/auth/change-password', data: {
        'current_password': currentPassword,
        'new_password': newPassword,
      });
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<UserDto> me() async {
    try {
      final resp = await _dio.get('/auth/me');
      return UserDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(ref.watch(apiClientProvider), ref.watch(authStorageProvider));
});