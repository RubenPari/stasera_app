import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/models/dto.dart';
import '../../providers/settings_provider.dart';

/// Sezione "Le mie preferenze": cibi da evitare, cucine preferite, tempo massimo.
///
/// Estratta dalla god-class `SettingsScreen`. Tutti gli update convergono nel
/// singolo metodo parametrico `PreferencesNotifier.updatePreferences`.
class PreferencesSection extends ConsumerStatefulWidget {
  const PreferencesSection({super.key});

  @override
  ConsumerState<PreferencesSection> createState() => _PreferencesSectionState();
}

class _PreferencesSectionState extends ConsumerState<PreferencesSection> {
  final _newDislikedCtrl = TextEditingController();
  final _newCuisineCtrl = TextEditingController();
  int? _pendingMaxPrep;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => ref.read(preferencesProvider.notifier).load(),
    );
  }

  @override
  void dispose() {
    _newDislikedCtrl.dispose();
    _newCuisineCtrl.dispose();
    super.dispose();
  }

  Future<void> _addDisliked(PreferencesDto p) async {
    final name = _newDislikedCtrl.text.trim();
    if (name.isEmpty) return;
    _newDislikedCtrl.clear();
    final err = await ref.read(preferencesProvider.notifier).updatePreferences(
      disliked: [...p.dislikedIngredients, name],
    );
    _reportError(err);
  }

  Future<void> _removeDisliked(PreferencesDto p, String name) async {
    final err = await ref.read(preferencesProvider.notifier).updatePreferences(
          disliked: p.dislikedIngredients.where((e) => e != name).toList(),
        );
    _reportError(err);
  }

  Future<void> _addCuisine(PreferencesDto p) async {
    final name = _newCuisineCtrl.text.trim();
    if (name.isEmpty) return;
    _newCuisineCtrl.clear();
    final err = await ref.read(preferencesProvider.notifier).updatePreferences(
      cuisines: [...p.preferredCuisines, name],
    );
    _reportError(err);
  }

  Future<void> _removeCuisine(PreferencesDto p, String name) async {
    final err = await ref.read(preferencesProvider.notifier).updatePreferences(
          cuisines: p.preferredCuisines.where((e) => e != name).toList(),
        );
    _reportError(err);
  }

  Future<void> _commitMaxPrep(PreferencesDto p, int value) async {
    final err = await ref.read(preferencesProvider.notifier).updatePreferences(
          maxPrep: value,
        );
    if (mounted) setState(() => _pendingMaxPrep = null);
    _reportError(err);
  }

  void _reportError(String? err) {
    if (err == null || !mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Errore: $err')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(preferencesProvider);

    if (state is PreferencesLoading || state is PreferencesInitial) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is PreferencesError) {
      return Text('Errore: ${state.message}');
    }
    final p = (state as PreferencesLoaded).prefs;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Cibi da evitare'),
            Wrap(
              spacing: 8,
              children: p.dislikedIngredients
                  .map((d) => Chip(
                        label: Text(d),
                        onDeleted: () => _removeDisliked(p, d),
                      ),)
                  .toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newDislikedCtrl,
                    decoration: const InputDecoration(
                        hintText: 'Aggiungi cibo da evitare',),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addDisliked(p),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Text('Cucine preferite'),
            Wrap(
              spacing: 8,
              children: p.preferredCuisines
                  .map((c) => Chip(
                        label: Text(c),
                        onDeleted: () => _removeCuisine(p, c),
                      ),)
                  .toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newCuisineCtrl,
                    decoration:
                        const InputDecoration(hintText: 'Aggiungi cucina'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addCuisine(p),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text('Tempo massimo: ${_pendingMaxPrep ?? p.maxPrepMinutes} min'),
            Slider(
              min: 5,
              max: 120,
              divisions: 23,
              value: ((_pendingMaxPrep ?? p.maxPrepMinutes).clamp(5, 120))
                  .toDouble(),
              label: (_pendingMaxPrep ?? p.maxPrepMinutes).toString(),
              onChanged: (v) => setState(() => _pendingMaxPrep = v.round()),
              onChangeEnd: (v) => _commitMaxPrep(p, v.round()),
            ),
          ],
        ),
      ),
    );
  }
}
