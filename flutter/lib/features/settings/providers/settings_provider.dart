import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../data/settings_repository.dart';

/// Provider asincrono per le preferenze utente.
final preferencesProvider =
    FutureProvider<PreferencesDto>((ref) async {
  final repo = ref.watch(settingsRepositoryProvider);
  return repo.getPreferences();
});

/// Provider asincrono per la lista staples.
final staplesProvider =
    FutureProvider<List<StapleDto>>((ref) async {
  final repo = ref.watch(settingsRepositoryProvider);
  return repo.listStaples();
});