import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Wrapper sicuro per la persistenza dei token JWT.
/// Usa flutter_secure_storage (Keychain iOS / EncryptedSharedPreferences Android).
class AuthStorage {
  AuthStorage(this._storage);

  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';

  final FlutterSecureStorage _storage;

  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    await _storage.write(key: _accessTokenKey, value: accessToken);
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  Future<String?> getAccessToken() => _storage.read(key: _accessTokenKey);
  Future<String?> getRefreshToken() => _storage.read(key: _refreshTokenKey);

  Future<bool> hasAccessToken() async {
    final token = await _storage.read(key: _accessTokenKey);
    return token != null && token.isNotEmpty;
  }

  Future<void> clear() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }
}

/// Provider singleton per AuthStorage.
final authStorageProvider = Provider<AuthStorage>((ref) {
  return AuthStorage(const FlutterSecureStorage());
});
