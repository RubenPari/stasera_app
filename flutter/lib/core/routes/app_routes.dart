/// Costanti di route centralizzate per go_router.
///
/// Evita string literal sparse nel codebase: ogni route dell'app è dichiarata
/// qui come `static const`. Per le route con parametri di path si espone sia
/// il template (usato in `GoRoute.path`) sia un factory che produce la
/// location concreta con i parametri interpolati.
class AppRoutes {
  AppRoutes._();

  /// Location di base dell'app (tab home).
  static const String home = '/';

  /// Schermata di cottura step-by-step.
  static const String cooking = '/cooking/:recipeId';

  /// Pianificatore settimanale.
  static const String week = '/week';

  /// Lista della spesa.
  static const String shopping = '/shopping';

  /// Modalità "troppo stanco" — 3 ricette di emergenza.
  static const String rescue = '/rescue';

  /// Impostazioni.
  static const String settings = '/settings';

  /// Login.
  static const String login = '/login';

  /// Registrazione.
  static const String register = '/register';

  /// Costruisce la location della schermata di cottura per una data ricetta.
  static String cookingFor(String recipeId) => '/cooking/$recipeId';
}
