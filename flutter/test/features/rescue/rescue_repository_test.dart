import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:stasera/features/rescue/data/rescue_repository.dart';

/// Stub HTTP adapter che risponde in base al predicato di match.
/// (Stesso pattern di test/features/auth/auth_repository_test.dart, niente
/// mock library: solo un adapter Dio che restituisce un body fisso.)
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
    final bytes = utf8.encode(jsonEncode(_body));
    return ResponseBody.fromBytes(
      bytes,
      _statusCode,
      headers: {
        'content-type': ['application/json'],
      },
    );
  }
}

void main() {
  group('RescueRepository.generate', () {
    test('parsa un BARE ARRAY di ricette (contratto post-refactor)', () async {
      final dio = Dio()
        ..httpClientAdapter = _StubAdapter(
          (o) => o.method == 'POST' && o.path == '/ai/rescue',
          200,
          [
            {
              'id': '11111111-1111-1111-1111-111111111111',
              'user_id': '22222222-2222-2222-2222-222222222222',
              'name': 'Pasta al tonno',
              'prep_minutes': 10,
              'servings': 1,
              'ingredients': [
                {'name': 'pasta', 'qty': '100g'},
              ],
              'steps': [
                {'text': 'cuoci', 'timer_seconds': 300},
              ],
              'is_rescue': true,
              'times_cooked': 0,
              'created_at': '2025-01-01T00:00:00Z',
            },
            {
              'id': '33333333-3333-3333-3333-333333333333',
              'user_id': '22222222-2222-2222-2222-222222222222',
              'name': 'Uova strapazzate',
              'prep_minutes': 5,
              'servings': 1,
              'ingredients': [],
              'steps': [],
              'is_rescue': true,
              'times_cooked': 0,
              'created_at': '2025-01-01T00:00:00Z',
            },
          ],
        );
      final repo = RescueRepository(dio);

      final recipes = await repo.generate();

      expect(recipes, hasLength(2));
      expect(recipes.first.name, 'Pasta al tonno');
      expect(recipes.first.isRescue, isTrue);
      expect(recipes.first.steps.single.timerSeconds, 300);
      expect(recipes.last.ingredients, isEmpty);
    });
  });
}