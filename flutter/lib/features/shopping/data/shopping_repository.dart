import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/dto.dart';

/// Repository per endpoint shopping-list.
class ShoppingRepository {
  ShoppingRepository(this._dio);

  final Dio _dio;

  Future<ShoppingListDto?> getCurrent() async {
    try {
      final resp = await _dio.get('/shopping-list/current');
      return ShoppingListDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      final api = toApiException(e);
      if (api.isNotFound) return null;
      throw api;
    }
  }

  Future<ShoppingListDto> generate() async {
    try {
      final resp = await _dio.post('/shopping-list/generate', data: {});
      return ShoppingListDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<ShoppingItemDto> updateItem(String itemId, bool isChecked) async {
    try {
      final resp = await _dio.patch(
        '/shopping-list/items/$itemId',
        data: {'is_checked': isChecked},
      );
      return ShoppingItemDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<ShoppingListDto> complete() async {
    try {
      final resp = await _dio.post('/shopping-list/current/complete', data: {});
      return ShoppingListDto.fromJson(resp.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }
}

final shoppingRepositoryProvider = Provider<ShoppingRepository>((ref) {
  return ShoppingRepository(ref.watch(apiClientProvider));
});