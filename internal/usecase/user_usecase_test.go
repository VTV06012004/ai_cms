package usecase

import (
	"context"
	"errors"
	"testing"

	"api/internal/domain"
)

type userRepositoryStub struct {
	users     map[uint]*domain.User
	nextID    uint
	lookupErr error
	updateErr error
}

func newUserRepositoryStub() *userRepositoryStub {
	return &userRepositoryStub{users: make(map[uint]*domain.User), nextID: 1}
}

func (r *userRepositoryStub) Create(_ context.Context, user *domain.User) (*domain.User, error) {
	data := user.Data()
	for _, existing := range r.users {
		if existing.Data().Email == data.Email {
			return nil, domain.ErrEmailAlreadyExists
		}
	}
	data.ID = r.nextID
	r.nextID++
	created := domain.ReconstituteUser(data)
	r.users[data.ID] = created
	return created, nil
}

func (r *userRepositoryStub) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	for _, user := range r.users {
		if string(user.Data().Email) == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *userRepositoryStub) GetByID(_ context.Context, id uint) (*domain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *userRepositoryStub) GetAll(_ context.Context, _ domain.PageRequest) ([]domain.User, int64, error) {
	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, int64(len(users)), nil
}

func (r *userRepositoryStub) CountByRole(_ context.Context, role domain.UserRole) (int64, error) {
	var count int64
	for _, user := range r.users {
		if user.Data().Role == role {
			count++
		}
	}
	return count, nil
}

func (r *userRepositoryStub) UpdateRole(_ context.Context, userID uint, newRole string) error {
	user, ok := r.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	return user.ChangeRole(domain.UserRole(newRole))
}

func (r *userRepositoryStub) UpdatePasswordHash(_ context.Context, userID uint, passwordHash string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	user, ok := r.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	data := user.Data()
	data.PasswordHash = passwordHash
	r.users[userID] = domain.ReconstituteUser(data)
	return nil
}

type auditRepositoryStub struct {
	actions []string
}

func (r *auditRepositoryStub) RecordLog(_ context.Context, _ uint, action, _, _, _ string) error {
	r.actions = append(r.actions, action)
	return nil
}

func (*auditRepositoryStub) GetAll(context.Context, domain.PageRequest) ([]domain.AuditLog, int64, error) {
	return nil, 0, nil
}

type unitOfWorkStub struct {
	users  domain.UserRepository
	audits domain.AuditRepository
}

func (u unitOfWorkStub) WithinTransaction(_ context.Context, work func(TransactionRepositories) error) error {
	return work(TransactionRepositories{Users: u.users, Audits: u.audits})
}

type credentialServiceStub struct{}

func (credentialServiceStub) HashPassword(password string) (string, error) {
	return "hashed:" + password, nil
}

func (credentialServiceStub) CheckPasswordHash(password, hash string) bool {
	return hash == "hashed:"+password
}

func (credentialServiceStub) PasswordHashNeedsUpgrade(string) bool {
	return false
}

func (credentialServiceStub) GenerateToken(uint, string, string) (string, error) {
	return "test-token", nil
}

type legacyCredentialServiceStub struct{}

func (legacyCredentialServiceStub) HashPassword(password string) (string, error) {
	return "argon2id:" + password, nil
}

func (legacyCredentialServiceStub) CheckPasswordHash(password, hash string) bool {
	return hash == "pbkdf2:"+password
}

func (legacyCredentialServiceStub) PasswordHashNeedsUpgrade(hash string) bool {
	return len(hash) >= len("pbkdf2:") && hash[:len("pbkdf2:")] == "pbkdf2:"
}

func (legacyCredentialServiceStub) GenerateToken(uint, string, string) (string, error) {
	return "test-token", nil
}

type countingCredentialServiceStub struct {
	hashCalls int
}

func (s *countingCredentialServiceStub) HashPassword(string) (string, error) {
	s.hashCalls++
	return "hashed", nil
}

func (*countingCredentialServiceStub) CheckPasswordHash(string, string) bool { return false }
func (*countingCredentialServiceStub) PasswordHashNeedsUpgrade(string) bool  { return false }
func (*countingCredentialServiceStub) GenerateToken(uint, string, string) (string, error) {
	return "test-token", nil
}

func newTestUserUseCase(repository *userRepositoryStub, audits *auditRepositoryStub) *UserUseCase {
	return NewUserUseCase(repository, credentialServiceStub{}, unitOfWorkStub{
		users: repository, audits: audits,
	})
}

func TestLoginUpgradesLegacyPasswordHash(t *testing.T) {
	users := newUserRepositoryStub()
	legacyUser, err := domain.NewUser("Test User", "user@example.com", "pbkdf2:secret")
	if err != nil {
		t.Fatal(err)
	}
	created, err := users.Create(context.Background(), legacyUser)
	if err != nil {
		t.Fatal(err)
	}
	audits := &auditRepositoryStub{}
	useCase := NewUserUseCase(users, legacyCredentialServiceStub{}, unitOfWorkStub{
		users: users, audits: audits,
	})

	if _, err := useCase.Login(context.Background(), LoginDTO{
		Email: "user@example.com", Password: "secret",
	}, "127.0.0.1"); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if got := users.users[created.Data().ID].Data().PasswordHash; got != "argon2id:secret" {
		t.Fatalf("password hash = %q, want upgraded Argon2id hash", got)
	}
}

func TestRegisterValidatesNameAndEmailBeforeHashing(t *testing.T) {
	users := newUserRepositoryStub()
	credentials := &countingCredentialServiceStub{}
	useCase := NewUserUseCase(users, credentials, unitOfWorkStub{users: users})

	if _, err := useCase.Register(context.Background(), RegisterDTO{
		Name: " ", Email: "invalid", Password: "password",
	}, "127.0.0.1"); !errors.Is(err, domain.ErrUserNameRequired) {
		t.Fatalf("blank name error = %v, want ErrUserNameRequired", err)
	}
	if _, err := useCase.Register(context.Background(), RegisterDTO{
		Name: "Test User", Email: "invalid", Password: "password",
	}, "127.0.0.1"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("invalid email error = %v, want ErrInvalidEmail", err)
	}
	if credentials.hashCalls != 0 {
		t.Fatalf("HashPassword called %d times for invalid registration data", credentials.hashCalls)
	}
}

func TestEnsureInitialAdminCreatesProtectedAdminOnce(t *testing.T) {
	users := newUserRepositoryStub()
	audits := &auditRepositoryStub{}
	useCase := newTestUserUseCase(users, audits)
	config := InitialAdminConfig{
		Name: "Site Admin", Email: "admin@example.com", Password: "strong-password",
	}

	if err := useCase.EnsureInitialAdmin(context.Background(), config); err != nil {
		t.Fatalf("EnsureInitialAdmin returned error: %v", err)
	}
	if err := useCase.EnsureInitialAdmin(context.Background(), InitialAdminConfig{
		Email: "changed@example.com", Password: "different-password",
	}); err != nil {
		t.Fatalf("repeated EnsureInitialAdmin returned error: %v", err)
	}
	if len(users.users) != 1 {
		t.Fatalf("expected one user after repeated bootstrap, got %d", len(users.users))
	}
	admin := users.users[1].Data()
	if admin.Role != domain.UserRoleAdmin || !admin.RootAdmin {
		t.Fatalf("expected root admin, got role %q and root flag %t", admin.Role, admin.RootAdmin)
	}
	if admin.PasswordHash != "hashed:strong-password" {
		t.Fatal("bootstrap password was not hashed or was unexpectedly overwritten")
	}
	if len(audits.actions) != 1 || audits.actions[0] != "BOOTSTRAP_ADMIN" {
		t.Fatalf("unexpected bootstrap audit actions: %v", audits.actions)
	}
}

func TestEnsureInitialAdminRequiresCredentialsWhenNoAdminExists(t *testing.T) {
	useCase := newTestUserUseCase(newUserRepositoryStub(), &auditRepositoryStub{})
	err := useCase.EnsureInitialAdmin(context.Background(), InitialAdminConfig{})
	if err == nil {
		t.Fatal("expected missing initial admin credentials to fail")
	}
}

func TestEnsureInitialAdminRejectsShortPassword(t *testing.T) {
	useCase := newTestUserUseCase(newUserRepositoryStub(), &auditRepositoryStub{})
	err := useCase.EnsureInitialAdmin(context.Background(), InitialAdminConfig{
		Email: "admin@example.com", Password: "short",
	})
	if err == nil {
		t.Fatal("expected short initial admin password to fail")
	}
}

func TestEnsureInitialAdminDoesNotRequireCredentialsWhenAdminExists(t *testing.T) {
	users := newUserRepositoryStub()
	admin, err := domain.NewUser("Existing Admin", "existing@example.com", "stored-hash")
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}
	if err := admin.ChangeRole(domain.UserRoleAdmin); err != nil {
		t.Fatalf("ChangeRole returned error: %v", err)
	}
	if _, err := users.Create(context.Background(), admin); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	useCase := newTestUserUseCase(users, &auditRepositoryStub{})
	if err := useCase.EnsureInitialAdmin(context.Background(), InitialAdminConfig{}); err != nil {
		t.Fatalf("existing admin should allow startup without bootstrap credentials: %v", err)
	}
	if len(users.users) != 1 {
		t.Fatalf("expected existing admin to remain unchanged, got %d users", len(users.users))
	}
}

func TestRegisterAlwaysCreatesAuthor(t *testing.T) {
	users := newUserRepositoryStub()
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	id, err := useCase.Register(context.Background(), RegisterDTO{
		Name: "New User", Email: "new@example.com", Password: "password",
	}, "")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	registered, err := users.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if registered.Data().Role != domain.UserRoleAuthor || registered.Data().RootAdmin {
		t.Fatalf("public registration elevated privileges: %+v", registered.Data())
	}
}

func TestUpdateUserRoleProtectsRootIdentityNotNumericID(t *testing.T) {
	users := newUserRepositoryStub()
	audits := &auditRepositoryStub{}
	ordinary, err := domain.NewUser("Ordinary", "ordinary@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	ordinaryData := ordinary.Data()
	ordinaryData.ID = 1
	users.users[1] = domain.ReconstituteUser(ordinaryData)
	users.nextID = 2
	root, err := domain.NewRootAdminUser("Root", "root@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	rootData := root.Data()
	rootData.ID = 2
	users.users[2] = domain.ReconstituteUser(rootData)

	useCase := newTestUserUseCase(users, audits)
	if err := useCase.UpdateUserRole(context.Background(), 9, 1, string(domain.UserRoleEditor), ""); err != nil {
		t.Fatalf("ordinary user ID 1 should not be treated as root: %v", err)
	}
	if err := useCase.UpdateUserRole(context.Background(), 9, 2, string(domain.UserRoleAuthor), ""); !errors.Is(err, domain.ErrCannotChangeRootAdmin) {
		t.Fatalf("expected root admin to be protected, got %v", err)
	}
	if got := users.users[1].Data().Role; got != domain.UserRoleEditor {
		t.Fatalf("ordinary user role = %q, want editor", got)
	}
	if got := users.users[2].Data().Role; got != domain.UserRoleAdmin {
		t.Fatalf("root admin role = %q, want admin", got)
	}
}

func TestEnsureInitialAdminRejectsEmailAlreadyUsedByAuthor(t *testing.T) {
	users := newUserRepositoryStub()
	author, err := domain.NewUser("Author", "reserved@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(context.Background(), author); err != nil {
		t.Fatal(err)
	}
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	err = useCase.EnsureInitialAdmin(context.Background(), InitialAdminConfig{
		Email: "reserved@example.com", Password: "strong-password",
	})
	if err == nil {
		t.Fatal("expected bootstrap to reject an email already used by an author")
	}
	if got := users.users[1].Data().Role; got != domain.UserRoleAuthor {
		t.Fatalf("existing author was promoted to %q during bootstrap", got)
	}
}

func TestLoginReturnsInvalidCredentialsOnlyWhenUserDoesNotExist(t *testing.T) {
	users := newUserRepositoryStub()
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	_, err := useCase.Login(context.Background(), LoginDTO{Email: "nobody@example.com", Password: "secret1"}, "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown email: expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginDoesNotReportDatabaseFailureAsWrongPassword(t *testing.T) {
	users := newUserRepositoryStub()
	users.lookupErr = errors.New("connection refused")
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	_, err := useCase.Login(context.Background(), LoginDTO{Email: "a@example.com", Password: "secret1"}, "")
	if err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected a real error for a database failure, got %v", err)
	}
}

func TestRegisterReportsDuplicateEmail(t *testing.T) {
	users := newUserRepositoryStub()
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})
	input := RegisterDTO{Name: "A", Email: "dup@example.com", Password: "secret1"}

	if _, err := useCase.Register(context.Background(), input, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := useCase.Register(context.Background(), input, ""); !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegisterRejectsShortPasswordBeforeHashing(t *testing.T) {
	users := newUserRepositoryStub()
	credentials := &countingCredentialServiceStub{}
	useCase := NewUserUseCase(users, credentials, unitOfWorkStub{users: users})

	_, err := useCase.Register(context.Background(), RegisterDTO{
		Name: "Test User", Email: "user@example.com", Password: "12345",
	}, "127.0.0.1")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Fatalf("short password error = %v, want ErrPasswordTooShort", err)
	}
	if credentials.hashCalls != 0 {
		t.Fatalf("HashPassword called %d times for a short password", credentials.hashCalls)
	}
}

func TestLoginSucceedsWhenPasswordHashUpgradeFails(t *testing.T) {
	users := newUserRepositoryStub()
	legacyUser, err := domain.NewUser("Test User", "user@example.com", "pbkdf2:secret")
	if err != nil {
		t.Fatal(err)
	}
	created, err := users.Create(context.Background(), legacyUser)
	if err != nil {
		t.Fatal(err)
	}
	users.updateErr = errors.New("db unavailable")
	useCase := NewUserUseCase(users, legacyCredentialServiceStub{}, unitOfWorkStub{
		users: users, audits: &auditRepositoryStub{},
	})

	if _, err := useCase.Login(context.Background(), LoginDTO{
		Email: "user@example.com", Password: "secret",
	}, "127.0.0.1"); err != nil {
		t.Fatalf("Login failed because of the hash upgrade: %v", err)
	}
	if got := users.users[created.Data().ID].Data().PasswordHash; got != "pbkdf2:secret" {
		t.Fatalf("password hash = %q, want legacy hash kept", got)
	}
}
