import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../data/settings_repository.dart';

// --- Preferences ---

sealed class PreferencesState {
  const PreferencesState();
}

class PreferencesInitial extends PreferencesState {
  const PreferencesInitial();
}

class PreferencesLoading extends PreferencesState {
  const PreferencesLoading();
}

class PreferencesLoaded extends PreferencesState {
  const PreferencesLoaded(this.prefs);
  final PreferencesDto prefs;
}

class PreferencesError extends PreferencesState {
  const PreferencesError(this.message);
  final String message;
}

/// Notifier asincrono per le preferenze utente.
///
/// Sostituisce il `FutureProvider` precedente allineandosi alla convenzione
/// StateNotifier + sealed states. Espone un **unico** metodo parametrico
/// [updatePreferences] che sostituisce i 4 metodi duplicati di update
/// (add/remove disliked, add/remove cuisine, slider max-prep): ogni sezione
/// passa solo i campi che cambiano.
class PreferencesNotifier extends StateNotifier<PreferencesState> {
  PreferencesNotifier(this._repo) : super(const PreferencesInitial());

  final SettingsRepository _repo;

  Future<void> load() async {
    state = const PreferencesLoading();
    try {
      final prefs = await _repo.getPreferences();
      state = PreferencesLoaded(prefs);
    } catch (e) {
      state = PreferencesError(e.toString());
    }
  }

  /// Aggiorna le preferenze: i parametri `null` vengono preservati dallo
  /// stato corrente (deve essere [PreferencesLoaded]). Unico punto di update
  /// per cibi evitati, cucine preferite e tempo massimo.
  Future<String?> updatePreferences({
    List<String>? disliked,
    int? maxPrep,
    List<String>? cuisines,
  }) async {
    final s = state;
    if (s is! PreferencesLoaded) return 'Preferenze non caricate';
    final next = s.prefs.copyWith(
      dislikedIngredients: disliked,
      maxPrepMinutes: maxPrep,
      preferredCuisines: cuisines,
    );
    try {
      final updated = await _repo.updatePreferences(
        disliked: next.dislikedIngredients,
        maxPrep: next.maxPrepMinutes,
        cuisines: next.preferredCuisines,
      );
      state = PreferencesLoaded(updated);
      return null;
    } on Object catch (e) {
      return e.toString();
    }
  }
}

final preferencesProvider =
    StateNotifierProvider.autoDispose<PreferencesNotifier, PreferencesState>(
        (ref) {
  return PreferencesNotifier(ref.watch(settingsRepositoryProvider));
});

// --- Staples ---

sealed class StaplesState {
  const StaplesState();
}

class StaplesInitial extends StaplesState {
  const StaplesInitial();
}

class StaplesLoading extends StaplesState {
  const StaplesLoading();
}

class StaplesLoaded extends StaplesState {
  const StaplesLoaded(this.staples);
  final List<StapleDto> staples;
}

class StaplesError extends StaplesState {
  const StaplesError(this.message);
  final String message;
}

/// Notifier asincrono per la lista staples.
class StaplesNotifier extends StateNotifier<StaplesState> {
  StaplesNotifier(this._repo) : super(const StaplesInitial());

  final SettingsRepository _repo;

  Future<void> load() async {
    state = const StaplesLoading();
    try {
      final staples = await _repo.listStaples();
      state = StaplesLoaded(staples);
    } catch (e) {
      state = StaplesError(e.toString());
    }
  }

  Future<void> createStaple(String name) async {
    try {
      await _repo.createStaple(name);
      await load();
    } catch (e) {
      state = StaplesError(e.toString());
    }
  }

  Future<void> updateStaple(String id, bool isActive) async {
    try {
      await _repo.updateStaple(id, isActive);
      await load();
    } catch (e) {
      state = StaplesError(e.toString());
    }
  }

  Future<void> deleteStaple(String id) async {
    try {
      await _repo.deleteStaple(id);
      await load();
    } catch (e) {
      state = StaplesError(e.toString());
    }
  }
}

final staplesProvider =
    StateNotifierProvider.autoDispose<StaplesNotifier, StaplesState>((ref) {
  return StaplesNotifier(ref.watch(settingsRepositoryProvider));
});
