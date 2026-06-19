// Modelli DTO Dart che rispecchiano le risposte JSON del backend Stasera.
// Tutti immutabili e con `fromJson` factory.

class UserDto {
  const UserDto({
    required this.id,
    required this.email,
    required this.displayName,
    required this.createdAt,
  });

  final String id;
  final String email;
  final String displayName;
  final DateTime createdAt;

  factory UserDto.fromJson(Map<String, dynamic> json) {
    return UserDto(
      id: json['id'] as String,
      email: json['email'] as String,
      displayName: json['display_name'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class AuthResponseDto {
  const AuthResponseDto({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });

  final UserDto user;
  final String accessToken;
  final String refreshToken;

  factory AuthResponseDto.fromJson(Map<String, dynamic> json) {
    return AuthResponseDto(
      user: UserDto.fromJson(json['user'] as Map<String, dynamic>),
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
    );
  }
}

class TokenPairDto {
  const TokenPairDto({
    required this.accessToken,
    required this.refreshToken,
  });

  final String accessToken;
  final String refreshToken;

  factory TokenPairDto.fromJson(Map<String, dynamic> json) {
    return TokenPairDto(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
    );
  }
}

class RecipeIngredientDto {
  const RecipeIngredientDto({required this.name, required this.qty});

  final String name;
  final String qty;

  factory RecipeIngredientDto.fromJson(Map<String, dynamic> json) {
    return RecipeIngredientDto(
      name: json['name'] as String,
      qty: json['qty'] as String,
    );
  }
}

class RecipeStepDto {
  const RecipeStepDto({required this.text, this.timerSeconds = 0});

  final String text;
  final int timerSeconds;

  factory RecipeStepDto.fromJson(Map<String, dynamic> json) {
    return RecipeStepDto(
      text: json['text'] as String,
      timerSeconds: (json['timer_seconds'] as num?)?.toInt() ?? 0,
    );
  }
}

class RecipeDto {
  const RecipeDto({
    required this.id,
    required this.userId,
    required this.name,
    required this.prepMinutes,
    required this.servings,
    required this.ingredients,
    required this.steps,
    required this.isRescue,
    required this.timesCooked,
    required this.createdAt,
    this.lastCookedAt,
  });

  final String id;
  final String userId;
  final String name;
  final int prepMinutes;
  final int servings;
  final List<RecipeIngredientDto> ingredients;
  final List<RecipeStepDto> steps;
  final bool isRescue;
  final int timesCooked;
  final DateTime createdAt;
  final DateTime? lastCookedAt;

  factory RecipeDto.fromJson(Map<String, dynamic> json) {
    return RecipeDto(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      prepMinutes: (json['prep_minutes'] as num).toInt(),
      servings: (json['servings'] as num).toInt(),
      ingredients: (json['ingredients'] as List<dynamic>)
          .map((e) => RecipeIngredientDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      steps: (json['steps'] as List<dynamic>)
          .map((e) => RecipeStepDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      isRescue: json['is_rescue'] as bool,
      timesCooked: (json['times_cooked'] as num).toInt(),
      createdAt: DateTime.parse(json['created_at'] as String),
      lastCookedAt: json['last_cooked_at'] != null
          ? DateTime.parse(json['last_cooked_at'] as String)
          : null,
    );
  }
}

class MealPlanDayDto {
  const MealPlanDayDto({
    required this.id,
    required this.planId,
    required this.dayOfWeek,
    required this.recipeId,
    this.recipe,
  });

  final String id;
  final String planId;
  final int dayOfWeek;
  final String recipeId;
  final RecipeDto? recipe;

  factory MealPlanDayDto.fromJson(Map<String, dynamic> json) {
    return MealPlanDayDto(
      id: json['id'] as String,
      planId: json['plan_id'] as String,
      dayOfWeek: (json['day_of_week'] as num).toInt(),
      recipeId: json['recipe_id'] as String,
      recipe: json['recipe'] != null
          ? RecipeDto.fromJson(json['recipe'] as Map<String, dynamic>)
          : null,
    );
  }
}

class MealPlanDto {
  const MealPlanDto({
    required this.id,
    required this.userId,
    required this.weekStart,
    required this.status,
    required this.days,
    required this.createdAt,
  });

  final String id;
  final String userId;
  final DateTime weekStart;
  final String status;
  final List<MealPlanDayDto> days;
  final DateTime createdAt;

  factory MealPlanDto.fromJson(Map<String, dynamic> json) {
    return MealPlanDto(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      weekStart: DateTime.parse(json['week_start'] as String),
      status: json['status'] as String,
      days: (json['days'] as List<dynamic>)
          .map((e) => MealPlanDayDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class ShoppingItemDto {
  const ShoppingItemDto({
    required this.id,
    required this.listId,
    required this.name,
    required this.quantity,
    required this.aisle,
    required this.isChecked,
    required this.sortOrder,
  });

  final String id;
  final String listId;
  final String name;
  final String quantity;
  final String aisle;
  final bool isChecked;
  final int sortOrder;

  factory ShoppingItemDto.fromJson(Map<String, dynamic> json) {
    return ShoppingItemDto(
      id: json['id'] as String,
      listId: json['list_id'] as String,
      name: json['name'] as String,
      quantity: json['quantity'] as String,
      aisle: json['aisle'] as String,
      isChecked: json['is_checked'] as bool,
      sortOrder: (json['sort_order'] as num).toInt(),
    );
  }

  ShoppingItemDto copyWith({bool? isChecked}) {
    return ShoppingItemDto(
      id: id,
      listId: listId,
      name: name,
      quantity: quantity,
      aisle: aisle,
      isChecked: isChecked ?? this.isChecked,
      sortOrder: sortOrder,
    );
  }
}

class ShoppingListDto {
  const ShoppingListDto({
    required this.id,
    required this.userId,
    required this.items,
    required this.createdAt,
    this.planId,
    this.completedAt,
  });

  final String id;
  final String userId;
  final String? planId;
  final List<ShoppingItemDto> items;
  final DateTime createdAt;
  final DateTime? completedAt;

  factory ShoppingListDto.fromJson(Map<String, dynamic> json) {
    return ShoppingListDto(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      planId: json['plan_id'] as String?,
      items: (json['items'] as List<dynamic>)
          .map((e) => ShoppingItemDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      createdAt: DateTime.parse(json['created_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }
}

class StapleDto {
  const StapleDto({
    required this.id,
    required this.userId,
    required this.name,
    required this.isActive,
  });

  final String id;
  final String userId;
  final String name;
  final bool isActive;

  factory StapleDto.fromJson(Map<String, dynamic> json) {
    return StapleDto(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      isActive: json['is_active'] as bool,
    );
  }
}

class PreferencesDto {
  const PreferencesDto({
    required this.userId,
    required this.dislikedIngredients,
    required this.maxPrepMinutes,
    required this.preferredCuisines,
    required this.updatedAt,
  });

  final String userId;
  final List<String> dislikedIngredients;
  final int maxPrepMinutes;
  final List<String> preferredCuisines;
  final DateTime updatedAt;

  factory PreferencesDto.fromJson(Map<String, dynamic> json) {
    return PreferencesDto(
      userId: json['user_id'] as String,
      dislikedIngredients:
          (json['disliked_ingredients'] as List<dynamic>).cast<String>(),
      maxPrepMinutes: (json['max_prep_minutes'] as num).toInt(),
      preferredCuisines:
          (json['preferred_cuisines'] as List<dynamic>).cast<String>(),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }
}