import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../home/data/meal_plan_repository.dart';

/// Ricette cachate (non-rescue) mostrate nel selettore "Scegli ricetta".
final recipePickerProvider = FutureProvider<List<RecipeDto>>((ref) async {
  final repo = ref.watch(mealPlanRepositoryProvider);
  return repo.listRecipes(isRescue: false);
});
