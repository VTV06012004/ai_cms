package usecase

import (
	"errors"
	"testing"

	"api/internal/domain"
)

type userRepositoryStub struct {
	users  map[uint]*domain.User
	nextID uint
}

func newUserRepositoryStub() *userRepositoryStub {
	return &userRepositoryStub{users: make(map[uint]*domain.User), nextID: 1}
}

func (r *userRepositoryStub) Create(user *domain.User) (*domain.User, error) {
	data := user.Data()
	for _, existing := range r.users {
		if existing.Data().Email == data.Email {
			return nil, errors.New("duplicate email")
		}
	}
	data.ID = r.nextID
	r.nextID++
	created := domain.ReconstituteUser(data)
	r.users[data.ID] = created
	return created, nil
}

func (r *userRepositoryStub) GetByEmail(email string) (*domain.User, error) {
	for _, user := range r.users {
		if string(user.Data().Email) == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *userRepositoryStub) GetByID(id uint) (*domain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *userRepositoryStub) GetAll() ([]domain.User, error) {
	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, nil
}

func (r *userRepositoryStub) CountByRole(role domain.UserRole) (int64, error) {
	var count int64
	for _, user := range r.users {
		if user.Data().Role == role {
			count++
		}
	}
	return count, nil
}

func (r *userRepositoryStub) UpdateRole(userID uint, newRole string) error {
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	return user.ChangeRole(domain.UserRole(newRole))
}

type auditRepositoryStub struct {
	actions []string
}

func (r *auditRepositoryStub) RecordLog(_ uint, action, _, _, _ string) error {
	r.actions = append(r.actions, action)
	return nil
}

func (*auditRepositoryStub) GetAll() ([]domain.AuditLog, error) {
	return nil, nil
}

type unitOfWorkStub struct {
	users  domain.UserRepository
	audits domain.AuditRepository
}

func (u unitOfWorkStub) WithinTransaction(work func(TransactionRepositories) error) error {
	return work(TransactionRepositories{Users: u.users, Audits: u.audits})
}

type credentialServiceStub struct{}

func (credentialServiceStub) HashPassword(password string) (string, error) {
	return "hashed:" + password, nil
}

func (credentialServiceStub) CheckPasswordHash(password, hash string) bool {
	return hash == "hashed:"+password
}

func (credentialServiceStub) GenerateToken(uint, string, string) (string, error) {
	return "test-token", nil
}

func newTestUserUseCase(repository *userRepositoryStub, audits *auditRepositoryStub) *UserUseCase {
	return NewUserUseCase(repository, credentialServiceStub{}, unitOfWorkStub{
		users: repository, audits: audits,
	})
}

func TestEnsureInitialAdminCreatesProtectedAdminOnce(t *testing.T) {
	users := newUserRepositoryStub()
	audits := &auditRepositoryStub{}
	useCase := newTestUserUseCase(users, audits)
	config := InitialAdminConfig{
		Name: "Site Admin", Email: "admin@example.com", Password: "strong-password",
	}

	if err := useCase.EnsureInitialAdmin(config); err != nil {
		t.Fatalf("EnsureInitialAdmin returned error: %v", err)
	}
	if err := useCase.EnsureInitialAdmin(InitialAdminConfig{
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
	err := useCase.EnsureInitialAdmin(InitialAdminConfig{})
	if err == nil {
		t.Fatal("expected missing initial admin credentials to fail")
	}
}

func TestEnsureInitialAdminRejectsShortPassword(t *testing.T) {
	useCase := newTestUserUseCase(newUserRepositoryStub(), &auditRepositoryStub{})
	err := useCase.EnsureInitialAdmin(InitialAdminConfig{
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
	if _, err := users.Create(admin); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	useCase := newTestUserUseCase(users, &auditRepositoryStub{})
	if err := useCase.EnsureInitialAdmin(InitialAdminConfig{}); err != nil {
		t.Fatalf("existing admin should allow startup without bootstrap credentials: %v", err)
	}
	if len(users.users) != 1 {
		t.Fatalf("expected existing admin to remain unchanged, got %d users", len(users.users))
	}
}

func TestRegisterAlwaysCreatesAuthor(t *testing.T) {
	users := newUserRepositoryStub()
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	id, err := useCase.Register(RegisterDTO{
		Name: "New User", Email: "new@example.com", Password: "password",
	}, "")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	registered, err := users.GetByID(id)
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
	if err := useCase.UpdateUserRole(9, 1, string(domain.UserRoleEditor), ""); err != nil {
		t.Fatalf("ordinary user ID 1 should not be treated as root: %v", err)
	}
	if err := useCase.UpdateUserRole(9, 2, string(domain.UserRoleAuthor), ""); !errors.Is(err, domain.ErrCannotChangeRootAdmin) {
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
	if _, err := users.Create(author); err != nil {
		t.Fatal(err)
	}
	useCase := newTestUserUseCase(users, &auditRepositoryStub{})

	err = useCase.EnsureInitialAdmin(InitialAdminConfig{
		Email: "reserved@example.com", Password: "strong-password",
	})
	if err == nil {
		t.Fatal("expected bootstrap to reject an email already used by an author")
	}
	if got := users.users[1].Data().Role; got != domain.UserRoleAuthor {
		t.Fatalf("existing author was promoted to %q during bootstrap", got)
	}
}
