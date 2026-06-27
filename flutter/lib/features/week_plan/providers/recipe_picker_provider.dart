import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../home/data/recipe_repository.dart';

/// Stato asincrono del selettore "Scegli ricetta".
sealed class RecipePickerState {
  const RecipePickerState();
}

class RecipePickerInitial extends RecipePickerState {
  const RecipePickerInitial();
}

class RecipePickerLoading extends RecipePickerState {
  const RecipePickerLoading();
}

class RecipePickerLoaded extends RecipePickerState {
  const RecipePickerLoaded(this.recipes);
  final List<RecipeDto> recipes;
}

class RecipePickerError extends RecipePickerState {
  const RecipePickerError(this.message);
  final String message;
}

/// Notifier asincrono per le ricette cachate (non-rescue) mostrate nel
/// selettore "Scegli ricetta". Sostituisce il `FutureProvider` precedente,
/// allineandosi alla convenzione StateNotifier + sealed states.
class RecipePickerNotifier extends StateNotifier<RecipePickerState> {
  RecipePickerNotifier(this._repo) : super(const RecipePickerInitial());

  final RecipeRepository _repo;

  Future<void> load() async {
    state = const RecipePickerLoading();
    try {
      final recipes = await _repo.listRecipes(isRescue: false);
      state = RecipePickerLoaded(recipes);
    } catch (e) {
      state = RecipePickerError(e.toString());
    }
  }
}

final recipePickerProvider =
    StateNotifierProvider.autoDispose<RecipePickerNotifier, RecipePickerState>(
        (ref) {
  return RecipePickerNotifier(ref.watch(recipeRepositoryProvider));
});
