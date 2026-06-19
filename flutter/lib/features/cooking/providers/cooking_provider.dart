import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../home/data/meal_plan_repository.dart';

/// Stato della sessione di cottura.
sealed class CookingState {
  const CookingState();
}

class CookingLoading extends CookingState {
  const CookingLoading();
}

class CookingReady extends CookingState {
  CookingReady(this.recipe, this.completedSteps, this.activeTimer);
  final RecipeDto recipe;
  final Set<int> completedSteps;
  final int? activeTimer; // step index in countdown
  bool get allDone => completedSteps.length == recipe.steps.length;
}

class CookingError extends CookingState {
  const CookingError(this.message);
  final String message;
}

class CookingDone extends CookingState {
  const CookingDone();
}

/// Notifier che gestisce step completati, timer attivi e chiamata POST /cooked.
class CookingNotifier extends StateNotifier<CookingState> {
  CookingNotifier(this._repo) : super(const CookingLoading());

  final MealPlanRepository _repo;
  Timer? _countdown;
  int _remainingSeconds = 0;

  /// Carica la ricetta per ID (via GET /recipes/:id).
  Future<void> load(String recipeId) async {
    try {
      final recipe = await _repo.getRecipe(recipeId);
      state = CookingReady(recipe, <int>{}, null);
    } catch (e) {
      state = CookingError(e.toString());
    }
  }

  void toggleStep(int index) {
    final s = state;
    if (s is! CookingReady) return;
    final completed = Set<int>.from(s.completedSteps);
    if (completed.contains(index)) {
      completed.remove(index);
    } else {
      completed.add(index);
    }
    state = CookingReady(s.recipe, completed, s.activeTimer);
  }

  void startTimer(int stepIndex, void Function() onTick, void Function() onDone) {
    final s = state;
    if (s is! CookingReady) return;
    final seconds = s.recipe.steps[stepIndex].timerSeconds;
    if (seconds <= 0) return;
    _remainingSeconds = seconds;
    state = CookingReady(s.recipe, s.completedSteps, stepIndex);

    _countdown?.cancel();
    _countdown = Timer.periodic(const Duration(seconds: 1), (t) {
      _remainingSeconds--;
      onTick();
      if (_remainingSeconds <= 0) {
        t.cancel();
        final cur = state;
        if (cur is CookingReady) {
          state = CookingReady(cur.recipe, cur.completedSteps, null);
        }
        onDone();
      }
    });
  }

  int get remainingSeconds => _remainingSeconds;

  void cancelTimer() {
    _countdown?.cancel();
    final s = state;
    if (s is CookingReady) {
      state = CookingReady(s.recipe, s.completedSteps, null);
    }
  }

  /// Completa la cottura, invia POST /recipes/:id/cooked.
  Future<void> finishCooking(String recipeId) async {
    _countdown?.cancel();
    try {
      await _repo.markCooked(recipeId);
      state = const CookingDone();
    } catch (e) {
      state = CookingError(e.toString());
    }
  }

  @override
  void dispose() {
    _countdown?.cancel();
    super.dispose();
  }
}

final cookingProvider =
    StateNotifierProvider.autoDispose<CookingNotifier, CookingState>((ref) {
  return CookingNotifier(ref.watch(mealPlanRepositoryProvider));
});