import '../../../core/models/dto.dart';

/// Sostituisce in [plan] il giorno il cui `dayOfWeek` coincide con quello di
/// [updatedDay], restituendo un nuovo [MealPlanDto] (immutabile).
///
/// Estratta dalla duplicazione tra `WeekPlanNotifier.swapDay` e
/// `regenerateDay`, che applicano entrambe la stessa sostituzione di giorno.
MealPlanDto replaceDayInPlan(MealPlanDto plan, MealPlanDayDto updatedDay) {
  final days = plan.days
      .map((d) => d.dayOfWeek == updatedDay.dayOfWeek ? updatedDay : d)
      .toList();
  return plan.copyWith(days: days);
}
