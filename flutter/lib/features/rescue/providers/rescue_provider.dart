import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../data/rescue_repository.dart';

sealed class RescueState {
  const RescueState();
}

class RescueInitial extends RescueState {
  const RescueInitial();
}

class RescueLoading extends RescueState {
  const RescueLoading();
}

class RescueLoaded extends RescueState {
  const RescueLoaded(this.recipes);
  final List<RecipeDto> recipes;
}

class RescueError extends RescueState {
  const RescueError(this.message);
  final String message;
}

class RescueNotifier extends StateNotifier<RescueState> {
  RescueNotifier(this._repo) : super(const RescueInitial());

  final RescueRepository _repo;

  Future<void> load() async {
    state = const RescueLoading();
    try {
      final recipes = await _repo.generate();
      state = RescueLoaded(recipes);
    } catch (e) {
      state = RescueError(e.toString());
    }
  }
}

final rescueProvider =
    StateNotifierProvider.autoDispose<RescueNotifier, RescueState>((ref) {
  return RescueNotifier(ref.watch(rescueRepositoryProvider));
});
