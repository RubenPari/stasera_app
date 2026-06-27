/// Eccezione lanciata dal client API per errori HTTP o di rete.
class ApiException implements Exception {
  ApiException({
    required this.statusCode,
    required this.message,
    this.error,
  });

  final int statusCode;
  final String message;
  final Object? error;

  bool get isUnauthorized => statusCode == 401;
  bool get isNotFound => statusCode == 404;
  bool get isConflict => statusCode == 409;
  bool get isValidation => statusCode == 400;

  @override
  String toString() => 'ApiException($statusCode): $message';
}
