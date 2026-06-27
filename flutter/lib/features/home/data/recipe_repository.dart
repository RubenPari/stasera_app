import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per gli endpoint `/recipes/*` (ricette cachate).
///
/// Separato da [MealPlanRepository] per responsabilità singola.
class RecipeRepository {
  RecipeRepository(this._dio);

  final Dio _dio;

  /// Dettaglio di una ricetta cachata.
  Future<RecipeDto> getRecipe(String id) async {
    try {
      final resp = await _dio.get('/recipes/$id');
      return RecipeDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Segna la ricetta come cucinata (incrementa times_cooked).
  Future<RecipeDto> markCooked(String id) async {
    try {
      final resp = await _dio.post('/recipes/$id/cooked');
      return RecipeDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Lista di tutte le ricette cachate, opzionalmente filtrate per is_rescue.
  Future<List<RecipeDto>> listRecipes({bool? isRescue}) async {
    try {
      final qs = isRescue == null ? '' : '?is_rescue=$isRescue';
      final resp = await _dio.get('/recipes$qs');
      return (resp.data as List<dynamic>)
          .map((e) => RecipeDto.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Elimina una ricetta cachata.
  Future<void> deleteRecipe(String id) async {
    try {
      await _dio.delete('/recipes/$id');
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final recipeRepositoryProvider = Provider<RecipeRepository>((ref) {
  return RecipeRepository(ref.watch(apiClientProvider));
});
