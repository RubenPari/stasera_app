import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'core/notifications/notification_service.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  NotificationService.init();
  runApp(
    const ProviderScope(
      child: StaseraApp(),
    ),
  );
}