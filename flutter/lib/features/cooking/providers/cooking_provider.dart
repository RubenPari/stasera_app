import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../home/data/recipe_repository.dart';

/// Stato della sessione di cottura.
sealed class CookingState {
  const CookingState();
}

class CookingLoading extends CookingState {
  const CookingLoading();
}

/// Stato "pronto": ricetta caricata, step completati, eventuale timer attivo.
///
/// Il countdown VIVE nel notifier: [activeTimerStep] è l'indice dello step in
/// countdown (null = nessun timer attivo), [remainingSeconds] è il conto alla
/// rovescia corrente. La UI si limita a fare il render di questi campi.
/// [justFinishedStep] è un segnale transiente (step il cui timer è appena
/// terminato) che la UI usa per mostrare l'overlay "Timer finito!"; va
/// consumato chiamando [CookingNotifier.clearJustFinished].
class CookingReady extends CookingState {
  CookingReady(
    this.recipe,
    this.completedSteps,
    this.activeTimerStep,
    this.remainingSeconds, {
    this.justFinishedStep,
  });

  final RecipeDto recipe;
  final Set<int> completedSteps;
  final int? activeTimerStep;
  final int remainingSeconds;
  final int? justFinishedStep;

  bool get allDone => completedSteps.length == recipe.steps.length;
}

class CookingError extends CookingState {
  const CookingError(this.message);
  final String message;
}

class CookingDone extends CookingState {
  const CookingDone();
}

/// Notifier che possiede tutta la logica di countdown e stato step.
/// La UI di [CookingScreen] si limita a fare il render di [CookingState].
class CookingNotifier extends StateNotifier<CookingState> {
  CookingNotifier(this._repo) : super(const CookingLoading());

  final RecipeRepository _repo;
  Timer? _countdown;

  /// Carica la ricetta per ID (via GET /recipes/:id).
  Future<void> load(String recipeId) async {
    try {
      final recipe = await _repo.getRecipe(recipeId);
      state = CookingReady(recipe, <int>{}, null, 0);
    } catch (e) {
      state = CookingError(e.toString());
    }
  }

  /// Toggle (on/off) di uno step senza timer — per i tap manuali.
  void toggleStep(int index) {
    final s = state;
    if (s is! CookingReady) return;
    final completed = Set<int>.from(s.completedSteps);
    if (completed.contains(index)) {
      completed.remove(index);
    } else {
      completed.add(index);
    }
    state = CookingReady(
      s.recipe,
      completed,
      s.activeTimerStep,
      s.remainingSeconds,
    );
  }

  /// Avvia il countdown di uno step con timer. Aggiorna [remainingSeconds]
  /// ogni secondo; a countdown esaurito marca lo step come completato ed
  /// emette [justFinishedStep] per l'overlay UI.
  void startStepTimer(int stepIndex) {
    final s = state;
    if (s is! CookingReady) return;
    final seconds = s.recipe.steps[stepIndex].timerSeconds;
    if (seconds <= 0) return;

    _countdown?.cancel();
    state = CookingReady(s.recipe, s.completedSteps, stepIndex, seconds);

    _countdown = Timer.periodic(const Duration(seconds: 1), (t) {
      final cur = state;
      if (cur is! CookingReady) {
        t.cancel();
        return;
      }
      final next = cur.remainingSeconds - 1;
      if (next <= 0) {
        t.cancel();
        final completed = Set<int>.from(cur.completedSteps)
          ..add(cur.activeTimerStep!);
        state = CookingReady(
          cur.recipe,
          completed,
          null,
          0,
          justFinishedStep: cur.activeTimerStep,
        );
      } else {
        state = CookingReady(
          cur.recipe,
          cur.completedSteps,
          cur.activeTimerStep,
          next,
        );
      }
    });
  }

  /// Annulla il timer attivo (es. l'utente lascia la schermata o tap manuale).
  void cancelStepTimer() {
    _countdown?.cancel();
    final s = state;
    if (s is CookingReady) {
      state = CookingReady(s.recipe, s.completedSteps, null, 0);
    }
  }

  /// Consuma il segnale transiente [CookingReady.justFinishedStep]
  /// (chiamato dalla UI dopo aver mostrato l'overlay "Timer finito!").
  void clearJustFinished() {
    final s = state;
    if (s is CookingReady && s.justFinishedStep != null) {
      state = CookingReady(
        s.recipe,
        s.completedSteps,
        s.activeTimerStep,
        s.remainingSeconds,
      );
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
  return CookingNotifier(ref.watch(recipeRepositoryProvider));
});
