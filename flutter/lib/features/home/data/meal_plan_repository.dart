import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import '../../../core/models/dto.dart';

/// Repository per endpoint meal-plan e recipes.
class MealPlanRepository {
  MealPlanRepository(this._dio);

  final Dio _dio;

  /// Piano della settimana corrente (o 404 se non esiste).
  Future<MealPlanDto?> getCurrent() async {
    try {
      final resp = await _dio.get('/meal-plan/current');
      return MealPlanDto.fromJson(resp.data as Map<String, dynamic>);
    } on ApiException catch (e) {
      if (e.isNotFound) return null;
      rethrow;
    } on DioException catch (e) {
      final api = toApiException(e);
      if (api.isNotFound) return null;
      throw api;
    }
  }

  /// Genera un nuovo piano per la prossima settimana via AI.
  Future<MealPlanDto> generate() async {
    try {
      final resp = await _dio.post('/meal-plan/generate', data: {});
      return MealPlanDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Scambia la ricetta di un giorno con una già cachata.
  Future<MealPlanDayDto> swapDay({
    required String planId,
    required int dayOfWeek,
    required String recipeId,
  }) async {
    try {
      final resp = await _dio.patch(
        '/meal-plan/$planId/days/$dayOfWeek',
        data: {'recipe_id': recipeId},
      );
      return MealPlanDayDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Rigenera la ricetta di un giorno via AI.
  Future<MealPlanDayDto> regenerateDay({
    required String planId,
    required int dayOfWeek,
  }) async {
    try {
      final resp = await _dio.patch(
        '/meal-plan/$planId/days/$dayOfWeek',
        data: {'regenerate': true},
      );
      return MealPlanDayDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Ricetta di stasera (basata sul giorno corrente).
  Future<RecipeDto?> getToday() async {
    try {
      final resp = await _dio.get('/meal-plan/today');
      return RecipeDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      final api = toApiException(e);
      if (api.isNotFound) return null;
      throw api;
    }
  }

  /// Dettaglio di una ricetta cachata.
  Future<RecipeDto> getRecipe(String id) async {
    try {
      final resp = await _dio.get('/recipes/$id');
      return RecipeDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Segna la ricetta come cucinata stasera (incrementa times_cooked).
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
}

final mealPlanRepositoryProvider = Provider<MealPlanRepository>((ref) {
  return MealPlanRepository(ref.watch(apiClientProvider));
});