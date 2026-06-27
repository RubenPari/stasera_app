import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per endpoint staples + preferences.
class SettingsRepository {
  SettingsRepository(this._dio);

  final Dio _dio;

  // --- Staples ---
  Future<List<StapleDto>> listStaples() async {
    try {
      final resp = await _dio.get('/staples');
      return (resp.data as List<dynamic>)
          .map((e) => StapleDto.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<StapleDto> createStaple(String name) async {
    try {
      final resp = await _dio.post('/staples', data: {'name': name});
      return StapleDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<StapleDto> updateStaple(String id, bool isActive) async {
    try {
      final resp =
          await _dio.patch('/staples/$id', data: {'is_active': isActive});
      return StapleDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<void> deleteStaple(String id) async {
    try {
      await _dio.delete('/staples/$id');
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  // --- Preferences ---
  Future<PreferencesDto> getPreferences() async {
    try {
      final resp = await _dio.get('/preferences');
      return PreferencesDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<PreferencesDto> updatePreferences({
    required List<String> disliked,
    required int maxPrep,
    required List<String> cuisines,
  }) async {
    try {
      final resp = await _dio.patch('/preferences', data: {
        'disliked_ingredients': disliked,
        'max_prep_minutes': maxPrep,
        'preferred_cuisines': cuisines,
      },);
      return PreferencesDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final settingsRepositoryProvider = Provider<SettingsRepository>((ref) {
  return SettingsRepository(ref.watch(apiClientProvider));
});
