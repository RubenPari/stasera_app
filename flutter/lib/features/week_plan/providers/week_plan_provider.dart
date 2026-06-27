import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../home/data/meal_plan_repository.dart';
import '../../home/domain/plan_ops.dart';

sealed class WeekPlanState {
  const WeekPlanState();
}

class WeekPlanInitial extends WeekPlanState {
  const WeekPlanInitial();
}

class WeekPlanLoading extends WeekPlanState {
  const WeekPlanLoading();
}

class WeekPlanLoaded extends WeekPlanState {
  const WeekPlanLoaded(this.plan);
  final MealPlanDto plan;
}

class WeekPlanEmpty extends WeekPlanState {
  const WeekPlanEmpty();
}

class WeekPlanError extends WeekPlanState {
  const WeekPlanError(this.message);
  final String message;
}

class WeekPlanNotifier extends StateNotifier<WeekPlanState> {
  WeekPlanNotifier(this._repo) : super(const WeekPlanInitial());

  final MealPlanRepository _repo;

  Future<void> load() async {
    state = const WeekPlanLoading();
    try {
      final plan = await _repo.getCurrent();
      if (plan == null) {
        state = const WeekPlanEmpty();
      } else {
        state = WeekPlanLoaded(plan);
      }
    } catch (e) {
      state = WeekPlanError(e.toString());
    }
  }

  Future<void> generate() async {
    state = const WeekPlanLoading();
    try {
      final plan = await _repo.generate();
      state = WeekPlanLoaded(plan);
    } catch (e) {
      state = WeekPlanError(e.toString());
    }
  }

  Future<void> swapDay(int dayOfWeek, String recipeId) async {
    final s = state;
    if (s is! WeekPlanLoaded) return;
    try {
      final updated = await _repo.swapDay(
        planId: s.plan.id,
        dayOfWeek: dayOfWeek,
        recipeId: recipeId,
      );
      state = WeekPlanLoaded(replaceDayInPlan(s.plan, updated));
    } catch (e) {
      state = WeekPlanError(e.toString());
    }
  }

  Future<void> regenerateDay(int dayOfWeek) async {
    final s = state;
    if (s is! WeekPlanLoaded) return;
    try {
      final updated = await _repo.regenerateDay(
        planId: s.plan.id,
        dayOfWeek: dayOfWeek,
      );
      state = WeekPlanLoaded(replaceDayInPlan(s.plan, updated));
    } catch (e) {
      state = WeekPlanError(e.toString());
    }
  }
}

final weekPlanProvider =
    StateNotifierProvider.autoDispose<WeekPlanNotifier, WeekPlanState>((ref) {
  return WeekPlanNotifier(ref.watch(mealPlanRepositoryProvider));
});
