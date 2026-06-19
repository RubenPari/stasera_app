import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/models/dto.dart';
import '../../auth/providers/auth_provider.dart';
import '../data/settings_repository.dart';
import '../providers/settings_provider.dart';

/// Schermata impostazioni: preferenze, staple, account.
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final _newStapleCtrl = TextEditingController();
  final _newDislikedCtrl = TextEditingController();
  final _newCuisineCtrl = TextEditingController();

  @override
  void dispose() {
    _newStapleCtrl.dispose();
    _newDislikedCtrl.dispose();
    _newCuisineCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final prefs = ref.watch(preferencesProvider);
    final staples = ref.watch(staplesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Impostazioni')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _sectionTitle('Le mie preferenze'),
          prefs.when(
            data: (p) => _preferencesSection(p),
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('Errore: $e'),
          ),
          const SizedBox(height: 24),
          _sectionTitle('Prodotti sempre in casa'),
          staples.when(
            data: (list) => _staplesSection(list),
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('Errore: $e'),
          ),
          const SizedBox(height: 24),
          _sectionTitle('Account'),
          _accountSection(),
        ],
      ),
    );
  }

  Widget _sectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        title,
        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
      ),
    );
  }

  Widget _preferencesSection(PreferencesDto p) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Cibi da evitare'),
            Wrap(
              spacing: 8,
              children: p.dislikedIngredients.map((d) => Chip(
                label: Text(d),
                onDeleted: () => _removeDisliked(p, d),
              )).toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newDislikedCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi cibo da evitare'),
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
              children: p.preferredCuisines.map((c) => Chip(
                label: Text(c),
                onDeleted: () => _removeCuisine(p, c),
              )).toList(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newCuisineCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi cucina'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () => _addCuisine(p),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text('Tempo massimo di preparazione (min): ${p.maxPrepMinutes}'),
            // Il valore è statico qui; per slider live servirebbe uno StateNotifier dedicato.
          ],
        ),
      ),
    );
  }

  Widget _staplesSection(List<StapleDto> list) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ...list.map((s) => SwitchListTile(
                  value: s.isActive,
                  title: Text(s.name),
                  onChanged: (v) async {
                    await ref.read(settingsRepositoryProvider).updateStaple(s.id, v);
                    ref.invalidate(staplesProvider);
                  },
                  secondary: IconButton(
                    icon: const Icon(Icons.delete_outline),
                    onPressed: () async {
                      await ref.read(settingsRepositoryProvider).deleteStaple(s.id);
                      ref.invalidate(staplesProvider);
                    },
                  ),
                )),
            const Divider(),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _newStapleCtrl,
                    decoration: const InputDecoration(hintText: 'Aggiungi staple'),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: () async {
                    if (_newStapleCtrl.text.trim().isEmpty) return;
                    await ref
                        .read(settingsRepositoryProvider)
                        .createStaple(_newStapleCtrl.text.trim());
                    _newStapleCtrl.clear();
                    ref.invalidate(staplesProvider);
                  },
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _accountSection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Dettagli account potrebbero essere caricati da authProvider.
            const Text('Account'),
            const SizedBox(height: 8),
            FilledButton.tonalIcon(
              onPressed: () async {
                await ref.read(authProvider.notifier).logout();
              },
              icon: const Icon(Icons.logout),
              label: const Text('Esci'),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _addDisliked(PreferencesDto p) async {
    final name = _newDislikedCtrl.text.trim();
    if (name.isEmpty) return;
    final list = [...p.dislikedIngredients, name];
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: list,
          maxPrep: p.maxPrepMinutes,
          cuisines: p.preferredCuisines,
        );
    _newDislikedCtrl.clear();
    ref.invalidate(preferencesProvider);
  }

  Future<void> _removeDisliked(PreferencesDto p, String name) async {
    final list = p.dislikedIngredients.where((e) => e != name).toList();
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: list,
          maxPrep: p.maxPrepMinutes,
          cuisines: p.preferredCuisines,
        );
    ref.invalidate(preferencesProvider);
  }

  Future<void> _addCuisine(PreferencesDto p) async {
    final name = _newCuisineCtrl.text.trim();
    if (name.isEmpty) return;
    final list = [...p.preferredCuisines, name];
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: p.dislikedIngredients,
          maxPrep: p.maxPrepMinutes,
          cuisines: list,
        );
    _newCuisineCtrl.clear();
    ref.invalidate(preferencesProvider);
  }

  Future<void> _removeCuisine(PreferencesDto p, String name) async {
    final list = p.preferredCuisines.where((e) => e != name).toList();
    await ref.read(settingsRepositoryProvider).updatePreferences(
          disliked: p.dislikedIngredients,
          maxPrep: p.maxPrepMinutes,
          cuisines: list,
        );
    ref.invalidate(preferencesProvider);
  }
}