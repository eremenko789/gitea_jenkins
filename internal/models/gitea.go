package models

import "time"

// PullRequestEvent представляет событие создания Pull Request в Gitea
type PullRequestEvent struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}

// PullRequest представляет Pull Request
type PullRequest struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	State    string `json:"state"`
	URL      string `json:"html_url"`
	Head     Branch `json:"head"`
	Base     Branch `json:"base"`
	User     User   `json:"user"`
	Assignee *User  `json:"assignee"`
	Labels   []Label `json:"labels"`
}

// Branch представляет ветку
type Branch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo Repository `json:"repo"`
}

// Repository представляет репозиторий
type Repository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	URL         string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	SSHURL      string `json:"ssh_url"`
	Owner       User   `json:"owner"`
}

// User представляет пользователя
type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// Label представляет метку
type Label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Comment представляет комментарий
type Comment struct {
	ID      int    `json:"id"`
	Body    string `json:"body"`
	User    User   `json:"user"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
}

// CreateCommentRequest представляет запрос на создание комментария
type CreateCommentRequest struct {
	Body string `json:"body"`
}