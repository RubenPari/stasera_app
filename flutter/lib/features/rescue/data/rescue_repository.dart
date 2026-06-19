import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per endpoint AI rescue.
class RescueRepository {
  RescueRepository(this._dio);

  final Dio _dio;

  /// Genera 3 ricette di emergenza usando gli staple attivi.
  Future<List<RecipeDto>> generate() async {
    try {
      final resp = await _dio.post('/ai/rescue', data: {});
      final data = resp.data as Map<String, dynamic>;
      final recipes = (data['recipes'] as List<dynamic>)
          .map((e) => RecipeDto.fromJson(e as Map<String, dynamic>))
          .toList();
      return recipes;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final rescueRepositoryProvider = Provider<RescueRepository>((ref) {
  return RescueRepository(ref.watch(apiClientProvider));
});