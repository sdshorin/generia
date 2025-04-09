package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"instagram-clone/pkg/logger"
	"instagram-clone/pkg/middleware"
	"instagram-clone/pkg/models"
	"instagram-clone/pkg/services"
	"go.uber.org/zap"
)

// PostHandler обработчик для работы с постами
type PostHandler struct {
	postService services.PostService
	validate    *validator.Validate
}

// NewPostHandler создает новый обработчик постов
func NewPostHandler(postService services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
		validate:    validator.New(),
	}
}

// CreatePost обрабатывает запрос на создание поста
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var input models.CreatePostRequest

	// Получаем ID пользователя из контекста запроса
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Декодируем JSON запрос
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Logger.Warn("Failed to decode create post request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Создаем пост
	post, err := h.postService.CreatePost(r.Context(), userID, &input)
	if err != nil {
		logger.Logger.Error("Failed to create post", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// GetPost обрабатывает запрос на получение поста
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Получаем ID поста из URL
	vars := mux.Vars(r)
	postID := vars["id"]

	// Получаем ID пользователя из контекста запроса (если есть)
	userID := middleware.GetUserIDFromContext(r.Context())

	// Получаем пост
	post, err := h.postService.GetPostByID(r.Context(), postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("postID", postID))
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// GetGlobalFeed обрабатывает запрос на получение глобальной ленты постов
func (h *PostHandler) GetGlobalFeed(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры пагинации
	limit, offset := getPaginationParams(r)

	// Получаем ID пользователя из контекста запроса (если есть)
	userID := middleware.GetUserIDFromContext(r.Context())

	// Получаем глобальную ленту
	posts, err := h.postService.GetGlobalFeed(r.Context(), userID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get global feed", zap.Error(err))
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
		"pagination": map[string]int{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserPosts обрабатывает запрос на получение постов пользователя
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из URL
	vars := mux.Vars(r)
	targetUserID := vars["user_id"]

	// Получаем параметры пагинации
	limit, offset := getPaginationParams(r)

	// Получаем ID пользователя из контекста запроса (если есть)
	requestUserID := middleware.GetUserIDFromContext(r.Context())

	// Получаем посты пользователя
	posts, err := h.postService.GetUserPosts(r.Context(), targetUserID, requestUserID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get user posts", zap.Error(err), zap.String("userID", targetUserID))
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
		"pagination": map[string]int{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// LikePost обрабатывает запрос на лайк поста
func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста запроса
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из URL
	vars := mux.Vars(r)
	postID := vars["id"]

	// Ставим лайк
	err := h.postService.LikePost(r.Context(), postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to like post", zap.Error(err), zap.String("postID", postID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// UnlikePost обрабатывает запрос на удаление лайка с поста
func (h *PostHandler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста запроса
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из URL
	vars := mux.Vars(r)
	postID := vars["id"]

	// Удаляем лайк
	err := h.postService.UnlikePost(r.Context(), postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to unlike post", zap.Error(err), zap.String("postID", postID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// AddComment обрабатывает запрос на добавление комментария
func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	var input models.CreateCommentRequest

	// Получаем ID пользователя из контекста запроса
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из URL
	vars := mux.Vars(r)
	postID := vars["id"]

	// Декодируем JSON запрос
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Logger.Warn("Failed to decode create comment request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Валидируем данные
	err = h.validate.Struct(input)
	if err != nil {
		logger.Logger.Warn("Validation error", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем комментарий
	comment, err := h.postService.CreateComment(r.Context(), postID, userID, &input)
	if err != nil {
		logger.Logger.Error("Failed to create comment", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments обрабатывает запрос на получение комментариев к посту
func (h *PostHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Получаем ID поста из URL
	vars := mux.Vars(r)
	postID := vars["id"]

	// Получаем параметры пагинации
	limit, offset := getPaginationParams(r)

	// Получаем комментарии
	comments, err := h.postService.GetPostComments(r.Context(), postID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get comments", zap.Error(err), zap.String("postID", postID))
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comments": comments,
		"pagination": map[string]int{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getPaginationParams получает параметры пагинации из запроса
func getPaginationParams(r *http.Request) (int, int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // По умолчанию
	offset := 0 // По умолчанию

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}