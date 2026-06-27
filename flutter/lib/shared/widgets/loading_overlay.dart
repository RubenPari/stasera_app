import 'package:flutter/material.dart';

import '../../../core/theme/app_theme.dart';

/// Overlay a tutto schermo con spinner e messaggio.
///
/// Posizionato via `Stack` nelle schermate che lo usano (spesa, settimana):
/// uniformato a `AppTheme.scrim` / `AppTheme.onScrim` per la palette.
class LoadingOverlay extends StatelessWidget {
  const LoadingOverlay({super.key, required this.message, this.visible = true});

  final String message;
  final bool visible;

  @override
  Widget build(BuildContext context) {
    if (!visible) return const SizedBox.shrink();

    return Positioned.fill(
      child: AbsorbPointer(
        child: Container(
          color: AppTheme.scrim,
          child: Center(
            child: Card(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const CircularProgressIndicator(),
                    const SizedBox(height: 16),
                    Text(
                      message,
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: AppTheme.onScrim),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
