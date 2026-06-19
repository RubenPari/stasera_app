import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../shared/widgets/recipe_card.dart';
import '../providers/tonight_provider.dart';

/// Schermata principale: la cena di stasera.
/// Aperta ogni sera dopo il lavoro — deve essere istantanea e chiarissima.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    // Carica la cena di stasera al primo build.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(tonightProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(tonightProvider);
    final today = DateTime.now();
    final isSunday = today.weekday == DateTime.sunday;
    final isWeekday = today.weekday >= DateTime.monday &&
        today.weekday <= DateTime.friday;

    return Scaffold(
      appBar: AppBar(
        title: Column(
          children: [
            Text(_weekdayLabel(today.weekday)),
            Text(
              DateFormat('d MMMM', 'it_IT').format(today),
              style: const TextStyle(fontSize: 13),
            ),
          ],
        ),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.read(tonightProvider.notifier).load(),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: _body(state, isSunday, isWeekday),
        ),
      ),
    );
  }

  Widget _body(TonightState state, bool isSunday, bool isWeekday) {
    if (state is TonightLoading || state is TonightInitial) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is TonightError) {
      return Center(
        child: Text('Errore: ${state.message}',
            textAlign: TextAlign.center),
      );
    }
    if (state is TonightEmpty) {
      if (isSunday) {
        return _sundayEmpty();
      }
      if (!isWeekday) {
        return const Center(
          child: Text(
            'Oggi è weekend — nessuna cena pianificata.\n'
            'Buon riposo!',
            textAlign: TextAlign.center,
          ),
        );
      }
      return _weekdayEmpty();
    }
    if (state is TonightLoaded) {
      final recipe = state.recipe;
      return ListView(
        children: [
          const SizedBox(height: 8),
          Text(
            'Stasera cucini:',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 12),
          RecipeCard(
            recipe: recipe,
            onTap: () => context.push('/cooking/${recipe.id}'),
          ),
          const SizedBox(height: 24),
          FilledButton.icon(
            onPressed: () => context.push('/cooking/${recipe.id}'),
            icon: const Icon(Icons.play_arrow),
            label: const Text('Inizia a cucinare'),
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: () => context.push('/rescue'),
            icon: const Icon(Icons.bolt),
            label: const Text('Troppo stanco — cambia'),
          ),
        ],
      );
    }
    return const SizedBox.shrink();
  }

  Widget _sundayEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.calendar_today, size: 64),
          const SizedBox(height: 16),
          const Text(
            'È domenica! Pianifica le cene della settimana.',
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: () => context.push('/week'),
            child: const Text('Pianifica la settimana'),
          ),
        ],
      ),
    );
  }

  Widget _weekdayEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.nightlife, size: 64),
          const SizedBox(height: 16),
          const Text(
            'Nessuna cena pianificata per oggi.\n'
            'Vai su Settimana per generare il piano domenica.',
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: () => context.push('/week'),
            child: const Text('Apri Settimana'),
          ),
        ],
      ),
    );
  }

  String _weekdayLabel(int weekday) {
    const labels = {
      DateTime.monday: 'Lunedì',
      DateTime.tuesday: 'Martedì',
      DateTime.wednesday: 'Mercoledì',
      DateTime.thursday: 'Giovedì',
      DateTime.friday: 'Venerdì',
      DateTime.saturday: 'Sabato',
      DateTime.sunday: 'Domenica',
    };
    return labels[weekday] ?? '';
  }
}