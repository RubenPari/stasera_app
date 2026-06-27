import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../data/meal_plan_repository.dart';

/// Stato della cena di stasera.
sealed class TonightState {
  const TonightState();
}

class TonightInitial extends TonightState {
  const TonightInitial();
}

class TonightLoading extends TonightState {
  const TonightLoading();
}

class TonightLoaded extends TonightState {
  const TonightLoaded(this.recipe);
  final RecipeDto recipe;
}

class TonightEmpty extends TonightState {
  const TonightEmpty();
}

class TonightError extends TonightState {
  const TonightError(this.message);
  final String message;
}

class TonightNotifier extends StateNotifier<TonightState> {
  TonightNotifier(this._repo) : super(const TonightInitial());

  final MealPlanRepository _repo;

  Future<void> load() async {
    state = const TonightLoading();
    try {
      final recipe = await _repo.getToday();
      if (recipe == null) {
        state = const TonightEmpty();
      } else {
        state = TonightLoaded(recipe);
      }
    } catch (e) {
      state = TonightError(e.toString());
    }
  }
}

final tonightProvider =
    StateNotifierProvider.autoDispose<TonightNotifier, TonightState>((ref) {
  return TonightNotifier(ref.watch(mealPlanRepositoryProvider));
});
