import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lottie/lottie.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/util/format.dart';
import '../../../shared/util/load_on_post_frame.dart';
import '../../../shared/widgets/step_card.dart';
import '../providers/cooking_provider.dart';

/// Schermata di cottura step-by-step con timer individuali e animazione finale.
///
/// Tutta la logica di countdown vive in [CookingNotifier]; questa view si
/// limita a fare il render di [CookingState]. L'overlay "Timer finito!" è
/// driven da `ref.listen` sul segnale transiente `justFinishedStep`.
class CookingScreen extends ConsumerStatefulWidget {
  const CookingScreen({super.key, required this.recipeId});

  final String recipeId;

  @override
  ConsumerState<CookingScreen> createState() => _CookingScreenState();
}

class _CookingScreenState extends ConsumerState<CookingScreen> {
  bool _ingredientsExpanded = false;
  OverlayEntry? _timerOverlay;
  Timer? _overlayDismissTimer;

  @override
  void initState() {
    super.initState();
    loadOnPostFrame(
        () => ref.read(cookingProvider.notifier).load(widget.recipeId),);
    // Mostra l'overlay "Timer finito!" quando il notifier segnala il completamento.
    ref.listen<CookingState>(cookingProvider, (prev, next) {
      if (next is CookingReady && next.justFinishedStep != null) {
        _showTimerDoneOverlay();
        ref.read(cookingProvider.notifier).clearJustFinished();
      }
    });
  }

  @override
  void dispose() {
    _overlayDismissTimer?.cancel();
    _timerOverlay?.remove();
    super.dispose();
  }

  void _showTimerDoneOverlay() {
    _timerOverlay?.remove();
    _overlayDismissTimer?.cancel();
    _timerOverlay = OverlayEntry(
      builder: (ctx) => Material(
        color: AppTheme.scrim,
        child: Center(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.alarm, size: 64, color: AppTheme.success),
                  const SizedBox(height: 16),
                  const Text('Timer finito!', style: TextStyle(fontSize: 24)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => _dismissTimerOverlay(),
                    child: const Text('OK'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
    Overlay.of(context).insert(_timerOverlay!);
    // Auto-dismiss registrato per la dispose (non orfano).
    _overlayDismissTimer =
        Timer(const Duration(seconds: 10), _dismissTimerOverlay);
  }

  void _dismissTimerOverlay() {
    _overlayDismissTimer?.cancel();
    _overlayDismissTimer = null;
    _timerOverlay?.remove();
    _timerOverlay = null;
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(cookingProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Cucinando'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => context.pop(false),
        ),
      ),
      body: state is CookingLoading
          ? const Center(child: CircularProgressIndicator())
          : state is CookingError
              ? Center(child: Text('Errore: ${state.message}'))
              : state is CookingDone
                  ? _doneView()
                  : state is CookingReady
                      ? _readyView(state)
                      : const SizedBox.shrink(),
    );
  }

  Widget _readyView(CookingReady state) {
    final recipe = state.recipe;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(recipe.name, style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('${recipe.prepMinutes} min · ${recipe.servings} porzioni'),
        const SizedBox(height: 16),
        // Sezione ingredienti collassabile.
        Card(
          child: ExpansionTile(
            title: const Text('Ingredienti'),
            initiallyExpanded: _ingredientsExpanded,
            onExpansionChanged: (v) => setState(() => _ingredientsExpanded = v),
            children: recipe.ingredients
                .map((i) => ListTile(
                      dense: true,
                      title: Text(i.name),
                      trailing: Text(i.qty),
                    ),)
                .toList(),
          ),
        ),
        const SizedBox(height: 16),
        Text('Passaggi', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        ...recipe.steps.asMap().entries.map((entry) {
          final idx = entry.key;
          final step = entry.value;
          final isDone = state.completedSteps.contains(idx);
          final isActiveTimer = state.activeTimerStep == idx;
          return Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: StepCard(
              step: step,
              index: idx,
              isDone: isDone,
              timerLabel: isActiveTimer
                  ? Format.duration(state.remainingSeconds)
                  : null,
              onTap: () {
                if (step.timerSeconds > 0 && !isDone) {
                  ref.read(cookingProvider.notifier).startStepTimer(idx);
                } else {
                  ref.read(cookingProvider.notifier).toggleStep(idx);
                }
              },
            ),
          );
        }),
        const SizedBox(height: 24),
        if (state.allDone)
          FilledButton.icon(
            onPressed: () =>
                ref.read(cookingProvider.notifier).finishCooking(recipe.id),
            icon: const Icon(Icons.celebration),
            label: const Text('Finito!'),
          ),
      ],
    );
  }

  Widget _doneView() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Animazione Lottie confetti. Il file va in assets/lottie/confetti.json.
          // Se assente, fallback icona.
          Lottie.asset(
            'assets/lottie/confetti.json',
            repeat: false,
            width: 240,
            errorBuilder: (context, error, stack) => const Icon(
                Icons.celebration,
                size: 80,
                color: AppTheme.success,),
          ),
          const SizedBox(height: 16),
          const Text('Buona cena!', style: TextStyle(fontSize: 28)),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () => context.pop(true),
            child: const Text('Torna a casa'),
          ),
        ],
      ),
    );
  }
}
