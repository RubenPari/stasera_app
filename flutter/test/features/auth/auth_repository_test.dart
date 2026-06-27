import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:stasera/core/api/api_exception.dart';
import 'package:stasera/core/auth/auth_storage.dart';
import 'package:stasera/features/auth/data/auth_repository.dart';

/// Stub HTTP adapter che risponde in base al predicato di match.
class _StubAdapter implements HttpClientAdapter {
  _StubAdapter(this._match, this._statusCode, this._body);
  final bool Function(RequestOptions) _match;
  final int _statusCode;
  final Object _body;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (!_match(options)) {
      throw StateError('Unexpected request ${options.method} ${options.path}');
    }
    final bytes = utf8.encode(_body is String ? _body : jsonEncode(_body));
    return ResponseBody.fromBytes(
      bytes,
      _statusCode,
      headers: {
        'content-type': [_body is String ? 'text/plain' : 'application/json'],
      },
    );
  }
}

class _FakeStorage implements AuthStorage {
  @override
  Future<void> clear() async {}

  @override
  Future<String?> getAccessToken() async => null;

  @override
  Future<String?> getRefreshToken() async => null;

  @override
  Future<void> saveTokens({required String accessToken, required String refreshToken}) async {}

  @override
  Future<bool> hasAccessToken() async => false;
}

Dio _dio(bool Function(RequestOptions) m, int code, Object body) {
  final dio = Dio();
  dio.httpClientAdapter = _StubAdapter(m, code, body);
  return dio;
}

void main() {
  group('AuthRepository.updateMe', () {
    test('PATCH /auth/me ritorna il DTO aggiornato', () async {
      final dio = _dio(
        (o) => o.method == 'PATCH' && o.path == '/auth/me',
        200,
        {
          'id': '11111111-1111-1111-1111-111111111111',
          'email': 'v@e.com',
          'display_name': 'V2',
          'created_at': '2025-01-01T00:00:00Z',
        },
      );
      final repo = AuthRepository(dio, _FakeStorage());

      final user = await repo.updateMe(displayName: 'V2');

      expect(user.displayName, 'V2');
      expect(user.email, 'v@e.com');
    });

    test('PATCH /auth/me propaga ApiException su errore 400', () async {
      final dio = _dio(
        (o) => o.method == 'PATCH' && o.path == '/auth/me',
        400,
        {'error': 'display_name required'},
      );
      final repo = AuthRepository(dio, _FakeStorage());

      expect(
        repo.updateMe(displayName: ''),
        throwsA(isA<ApiException>()),
      );
    });
  });

  group('AuthRepository.changePassword', () {
    test('POST /auth/change-password accetta 204 senza body', () async {
      final dio = _dio(
        (o) => o.method == 'POST' && o.path == '/auth/change-password',
        204,
        '',
      );
      final repo = AuthRepository(dio, _FakeStorage());

      await repo.changePassword(
        currentPassword: 'password123',
        newPassword: 'newpass123',
      );
    });

    test('POST /auth/change-password 401 solleva ApiException 401', () async {
      final dio = _dio(
        (o) => o.method == 'POST' && o.path == '/auth/change-password',
        401,
        {'error': 'invalid current password'},
      );
      final repo = AuthRepository(dio, _FakeStorage());

      try {
        await repo.changePassword(
          currentPassword: 'sbagliata',
          newPassword: 'newpass123',
        );
        fail('attesa ApiException');
      } on ApiException catch (e) {
        expect(e.statusCode, 401);
      }
    });
  });
}
