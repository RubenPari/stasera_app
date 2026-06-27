import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per gli endpoint `/meal-plan/*` (piano settimanale).
///
/// Le chiamate alle ricette (`/recipes/*`) vivono in [RecipeRepository]:
/// il grab-bag precedente mescolava le due responsabilità.
class MealPlanRepository {
  MealPlanRepository(this._dio);

  final Dio _dio;

  /// Piano della settimana corrente (o null se 404).
  Future<MealPlanDto?> getCurrent() async {
    try {
      final resp = await _dio.get('/meal-plan/current');
      return MealPlanDto.fromJson(resp.data as Map<String, dynamic>);
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
    return _patchDay(
      planId: planId,
      dayOfWeek: dayOfWeek,
      body: {'recipe_id': recipeId},
    );
  }

  /// Rigenera la ricetta di un giorno via AI (riusando lo stesso endpoint
  /// PATCH con `regenerate: true`).
  Future<MealPlanDayDto> regenerateDay({
    required String planId,
    required int dayOfWeek,
  }) async {
    return _patchDay(
      planId: planId,
      dayOfWeek: dayOfWeek,
      body: {'regenerate': true},
    );
  }

  Future<MealPlanDayDto> _patchDay({
    required String planId,
    required int dayOfWeek,
    required Map<String, dynamic> body,
  }) async {
    try {
      final resp = await _dio.patch(
        '/meal-plan/$planId/days/$dayOfWeek',
        data: body,
      );
      return MealPlanDayDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  /// Ricetta di stasera (basata sul giorno corrente) o null se 404.
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
}

final mealPlanRepositoryProvider = Provider<MealPlanRepository>((ref) {
  return MealPlanRepository(ref.watch(apiClientProvider));
});
