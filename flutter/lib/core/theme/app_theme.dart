import 'package:flutter/material.dart';

/// Tema visivo dell'app Stasera.
/// Palette calda (rosso pomodoro + arancione emergenza) su sfondo neutro.
///
/// I colori semantici (success, warning, error, scrim, muted) vivono qui come
/// `static const` e vengono consumati dai widget al posto di `Colors.*`
/// hardcodati, in modo che la palette sia definita in un solo posto.
class AppTheme {
  AppTheme._();

  // --- Palette primaria ---
  static const Color primary = Color(0xFFC0392B);
  static const Color primaryDark = Color(0xFF922B21);
  static const Color accent = Color(0xFFE67E22);
  static const Color background = Color(0xFFF8F5F1);
  static const Color card = Colors.white;
  static const Color textMuted = Color(0xFF7F8C8D);

  // --- Colori semantici ---
  /// Stato positivo (es. "spesa completata", "timer finito").
  static const Color success = Color(0xFF27AE60);

  /// Messaggi di errore nei form (es. credenziali errate).
  static const Color error = Color(0xFFC0392B);

  /// Colore neutro per testo secondario / disabilitato.
  static const Color muted = Color(0xFF7F8C8D);

  /// Sfondo chiaro per avatar/elementi non completati.
  static const Color mutedSurface = Color(0xFFE0E0E0);

  /// Testo su superfici scure.
  static const Color onScrim = Colors.white;

  /// Testo principale su superfici chiare.
  static const Color onSurface = Colors.black87;

  /// Scrim semi-trasparente per overlay a tutto schermo.
  static const Color scrim = Colors.black54;

  static ThemeData get light {
    final base = ThemeData.light(useMaterial3: true);

    return base.copyWith(
      colorScheme: base.colorScheme.copyWith(
        primary: primary,
        secondary: accent,
        surface: card,
        error: error,
      ),
      scaffoldBackgroundColor: background,
      appBarTheme: const AppBarTheme(
        backgroundColor: primary,
        foregroundColor: onScrim,
        centerTitle: true,
        elevation: 0,
      ),
      cardTheme: CardThemeData(
        color: card,
        elevation: 2,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
        margin: EdgeInsets.zero,
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: primary,
          foregroundColor: onScrim,
          minimumSize: const Size.fromHeight(56),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          textStyle: const TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: accent,
          side: const BorderSide(color: accent, width: 2),
          minimumSize: const Size.fromHeight(56),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: card,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: muted),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: primary, width: 2),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 16,
        ),
      ),
    );
  }
}
