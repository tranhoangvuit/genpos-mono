package input

// SignUpInput carries validated SignUp parameters from the handler to the service.
type SignUpInput struct {
	Domain    string
	Email     string
	Password  string
	UserAgent string
}

// SignInInput carries validated SignIn parameters from the handler to the service.
type SignInInput struct {
	Email      string
	Password   string
	RememberMe bool
	UserAgent  string
}

// RefreshInput carries the refresh-token value extracted from the cookie.
type RefreshInput struct {
	RefreshToken string
	UserAgent    string
}

// SignOutInput carries the refresh-token value extracted from the cookie.
type SignOutInput struct {
	RefreshToken string
}
