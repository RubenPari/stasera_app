import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'core/auth/auth_storage.dart';
import 'core/routes/app_routes.dart';
import 'core/theme/app_theme.dart';
import 'features/auth/presentation/login_screen.dart';
import 'features/auth/presentation/register_screen.dart';
import 'features/cooking/presentation/cooking_screen.dart';
import 'features/home/presentation/home_screen.dart';
import 'features/rescue/presentation/rescue_screen.dart';
import 'features/settings/presentation/settings_screen.dart';
import 'features/shopping/presentation/shopping_screen.dart';
import 'features/week_plan/presentation/week_plan_screen.dart';

/// Provider che espone il router dell'app.
final routerProvider = Provider<GoRouter>((ref) {
  final storage = ref.watch(authStorageProvider);

  return GoRouter(
    initialLocation: AppRoutes.home,
    redirect: (context, state) async {
      final isLoggedIn = await storage.hasAccessToken();
      final isAuthRoute = state.matchedLocation == AppRoutes.login ||
          state.matchedLocation == AppRoutes.register;

      if (!isLoggedIn && !isAuthRoute) {
        return AppRoutes.login;
      }
      if (isLoggedIn && isAuthRoute) {
        return AppRoutes.home;
      }
      return null;
    },
    routes: [
      GoRoute(
        path: AppRoutes.home,
        builder: (context, state) => const HomeScreen(),
      ),
      GoRoute(
        path: AppRoutes.cooking,
        builder: (context, state) => CookingScreen(
          recipeId: state.pathParameters['recipeId']!,
        ),
      ),
      GoRoute(
        path: AppRoutes.week,
        builder: (context, state) => const WeekPlanScreen(),
      ),
      GoRoute(
        path: AppRoutes.shopping,
        builder: (context, state) => const ShoppingScreen(),
      ),
      GoRoute(
        path: AppRoutes.rescue,
        builder: (context, state) => const RescueScreen(),
      ),
      GoRoute(
        path: AppRoutes.settings,
        builder: (context, state) => const SettingsScreen(),
      ),
      GoRoute(
        path: AppRoutes.login,
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: AppRoutes.register,
        builder: (context, state) => const RegisterScreen(),
      ),
    ],
  );
});

/// Widget root dell'app: MaterialApp con tema e router Riverpod.
class StaseraApp extends ConsumerWidget {
  const StaseraApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Stasera',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      routerConfig: router,
    );
  }
}
