import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lottie/lottie.dart';

import '../../../shared/widgets/step_card.dart';
import '../providers/cooking_provider.dart';

/// Schermata di cottura step-by-step con timer individuali e animazione finale.
class CookingScreen extends ConsumerStatefulWidget {
  const CookingScreen({super.key, required this.recipeId});

  final String recipeId;

  @override
  ConsumerState<CookingScreen> createState() => _CookingScreenState();
}

class _CookingScreenState extends ConsumerState<CookingScreen> {
  bool _ingredientsExpanded = false;
  int? _activeStep;
  int _remaining = 0;
  Timer? _uiTimer;
  OverlayEntry? _timerOverlay;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(cookingProvider.notifier).load(widget.recipeId);
    });
  }

  @override
  void dispose() {
    _uiTimer?.cancel();
    _timerOverlay?.remove();
    super.dispose();
  }

  void _startTimerForStep(int stepIndex, int seconds) {
    if (seconds <= 0) return;
    setState(() => _activeStep = stepIndex);
    _remaining = seconds;

    _uiTimer?.cancel();
    _uiTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      setState(() => _remaining--);
      if (_remaining <= 0) {
        t.cancel();
        _showTimerDoneOverlay();
        ref.read(cookingProvider.notifier).toggleStep(stepIndex);
        setState(() => _activeStep = null);
      }
    });
  }

  void _showTimerDoneOverlay() {
    _timerOverlay = OverlayEntry(
      builder: (ctx) => Material(
        color: Colors.black54,
        child: Center(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.alarm, size: 64, color: Colors.green),
                  const SizedBox(height: 16),
                  const Text('Timer finito!',
                      style: TextStyle(fontSize: 24)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => _timerOverlay?.remove(),
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
    // Auto-dismiss dopo 10 secondi.
    Timer(const Duration(seconds: 10), () => _timerOverlay?.remove());
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
        Text(recipe.name,
            style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('${recipe.prepMinutes} min · ${recipe.servings} porzioni'),
        const SizedBox(height: 16),
        // Sezione ingredienti collassabile.
        Card(
          child: ExpansionTile(
            title: const Text('Ingredienti'),
            initiallyExpanded: _ingredientsExpanded,
            onExpansionChanged: (v) =>
                setState(() => _ingredientsExpanded = v),
            children: recipe.ingredients
                .map((i) => ListTile(
                      dense: true,
                      title: Text(i.name),
                      trailing: Text(i.qty),
                    ))
                .toList(),
          ),
        ),
        const SizedBox(height: 16),
        Text('Passaggi',
            style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        ...recipe.steps.asMap().entries.map((entry) {
          final idx = entry.key;
          final step = entry.value;
          final isDone = state.completedSteps.contains(idx);
          final isActiveTimer = _activeStep == idx;
          return Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: StepCard(
              step: step,
              index: idx,
              isDone: isDone,
              timerLabel: isActiveTimer ? _formatRemaining(_remaining) : null,
              onTap: () {
                if (step.timerSeconds > 0 && !isDone) {
                  _startTimerForStep(idx, step.timerSeconds);
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
            errorBuilder: (context, error, stack) =>
                const Icon(Icons.celebration, size: 80, color: Colors.green),
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

  String _formatRemaining(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }
}