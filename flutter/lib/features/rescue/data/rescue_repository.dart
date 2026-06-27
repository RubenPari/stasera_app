import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per l'endpoint AI rescue.
class RescueRepository {
  RescueRepository(this._dio);

  final Dio _dio;

  /// Genera 3 ricette di emergenza usando gli staple attivi.
  ///
  /// Il backend risponde con un **bare array** `[{recipe},{recipe},{recipe}]`
  /// (non più con envelope `{"recipes":[...]}`).
  Future<List<RecipeDto>> generate() async {
    try {
      final resp = await _dio.post('/ai/rescue', data: {});
      return (resp.data as List<dynamic>)
          .map((e) => RecipeDto.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final rescueRepositoryProvider = Provider<RescueRepository>((ref) {
  return RescueRepository(ref.watch(apiClientProvider));
});
