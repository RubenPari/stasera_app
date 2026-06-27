import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/models/dto.dart';
import '../../../core/routes/app_routes.dart';
import '../../../core/theme/app_theme.dart';
import '../../../shared/util/load_on_post_frame.dart';
import '../../../shared/widgets/loading_overlay.dart';
import '../providers/recipe_picker_provider.dart';
import '../providers/week_plan_provider.dart';

/// Pianificatore domenicale: mostra lun-ven, genera piano, rigenera giorno.
class WeekPlanScreen extends ConsumerStatefulWidget {
  const WeekPlanScreen({super.key});

  @override
  ConsumerState<WeekPlanScreen> createState() => _WeekPlanScreenState();
}

class _WeekPlanScreenState extends ConsumerState<WeekPlanScreen> {
  final _pickerSearchCtrl = TextEditingController();
  String _pickerQuery = '';

  @override
  void dispose() {
    _pickerSearchCtrl.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    loadOnPostFrame(() => ref.read(weekPlanProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(weekPlanProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Settimana')),
      body: Stack(
        children: [
          RefreshIndicator(
            onRefresh: () => ref.read(weekPlanProvider.notifier).load(),
            child: _body(state),
          ),
          LoadingOverlay(
            message: "L'AI sta pensando al tuo menù...",
            visible: state is WeekPlanLoading,
          ),
        ],
      ),
    );
  }

  Widget _body(WeekPlanState state) {
    if (state is WeekPlanInitial) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is WeekPlanError) {
      return Center(child: Text('Errore: ${state.message}'));
    }
    if (state is WeekPlanEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.auto_awesome, size: 64),
            const SizedBox(height: 16),
            const Text(
              'Nessun piano per questa settimana.',
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            FilledButton.icon(
              onPressed: () => ref.read(weekPlanProvider.notifier).generate(),
              icon: const Icon(Icons.auto_awesome),
              label: const Text('Genera piano con AI'),
            ),
          ],
        ),
      );
    }
    if (state is WeekPlanLoaded) {
      final plan = state.plan;
      return ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const Icon(Icons.event),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Settimana del ${DateFormat('d MMMM', 'it_IT').format(plan.weekStart)}',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          ...plan.days.map(_dayTile),
          const SizedBox(height: 24),
          FilledButton.icon(
            onPressed: () => context.push(AppRoutes.shopping),
            icon: const Icon(Icons.shopping_cart_outlined),
            label: const Text('Genera lista della spesa'),
          ),
        ],
      );
    }
    return const SizedBox.shrink();
  }

  Widget _dayTile(MealPlanDayDto day) {
    final recipe = day.recipe;
    final dayName = _dayName(day.dayOfWeek);

    return Card(
      child: ListTile(
        contentPadding: const EdgeInsets.all(16),
        title: Row(
          children: [
            Text(dayName, style: const TextStyle(fontWeight: FontWeight.bold)),
            const Spacer(),
            if (recipe != null)
              Text('${recipe.prepMinutes} min',
                  style: const TextStyle(color: AppTheme.muted),),
          ],
        ),
        subtitle: Text(recipe?.name ?? '—'),
        trailing: IconButton(
          icon: const Icon(Icons.swap_horiz),
          onPressed: () => _showDayOptions(day.dayOfWeek),
        ),
      ),
    );
  }

  void _showDayOptions(int dayOfWeek) {
    showModalBottomSheet<void>(
      context: context,
      builder: (ctx) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.refresh),
                title: const Text('Rigenera questo giorno'),
                onTap: () {
                  Navigator.pop(ctx);
                  ref.read(weekPlanProvider.notifier).regenerateDay(dayOfWeek);
                },
              ),
              ListTile(
                leading: const Icon(Icons.list),
                title: const Text('Scegli ricetta'),
                onTap: () {
                  Navigator.pop(ctx);
                  _openRecipePicker(dayOfWeek);
                },
              ),
            ],
          ),
        );
      },
    );
  }

  void _openRecipePicker(int dayOfWeek) {
    _pickerSearchCtrl.clear();
    _pickerQuery = '';
    // Carica le ricette al primo frame dell'apertura del picker.
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => ref.read(recipePickerProvider.notifier).load(),
    );
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) {
        return Padding(
          padding: EdgeInsets.only(
            bottom: MediaQuery.of(ctx).viewInsets.bottom,
          ),
          child: SafeArea(
            child: StatefulBuilder(
              builder: (ctx, setLocal) {
                final pickerState = ref.watch(recipePickerProvider);
                return Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 16, 8, 8),
                      child: Row(
                        children: [
                          Expanded(
                            child: TextField(
                              controller: _pickerSearchCtrl,
                              decoration: const InputDecoration(
                                hintText: 'Cerca ricetta',
                                prefixIcon: Icon(Icons.search),
                              ),
                              onChanged: (v) =>
                                  setLocal(() => _pickerQuery = v),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(Icons.close),
                            onPressed: () => Navigator.pop(ctx),
                          ),
                        ],
                      ),
                    ),
                    Flexible(child: _pickerBody(pickerState, dayOfWeek)),
                  ],
                );
              },
            ),
          ),
        );
      },
    );
  }

  Widget _pickerBody(RecipePickerState state, int dayOfWeek) {
    if (state is RecipePickerLoading || state is RecipePickerInitial) {
      return const Padding(
        padding: EdgeInsets.all(24),
        child: Center(child: CircularProgressIndicator()),
      );
    }
    if (state is RecipePickerError) {
      return Padding(
        padding: const EdgeInsets.all(24),
        child: Text('Errore: ${state.message}'),
      );
    }
    final recipes = (state as RecipePickerLoaded).recipes;
    final q = _pickerQuery.toLowerCase();
    final filtered = q.isEmpty
        ? recipes
        : recipes.where((r) => r.name.toLowerCase().contains(q)).toList();
    if (filtered.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(24),
        child: Text('Nessuna ricetta disponibile.'),
      );
    }
    return ListView.builder(
      shrinkWrap: true,
      itemCount: filtered.length,
      itemBuilder: (_, i) {
        final recipe = filtered[i];
        return ListTile(
          title: Text(recipe.name),
          subtitle: Text('${recipe.prepMinutes} min'),
          onTap: () {
            Navigator.pop(context);
            ref.read(weekPlanProvider.notifier).swapDay(dayOfWeek, recipe.id);
          },
        );
      },
    );
  }

  String _dayName(int dayOfWeek) {
    const names = {
      1: 'Lunedì',
      2: 'Martedì',
      3: 'Mercoledì',
      4: 'Giovedì',
      5: 'Venerdì',
    };
    return names[dayOfWeek] ?? '';
  }
}
