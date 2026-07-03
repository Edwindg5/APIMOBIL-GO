package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type RegisterService struct {
	usuarioRepository interfaces.UsuarioRepository
}

// NewRegisterService crea una nueva instancia del servicio de registro
func NewRegisterService(usuarioRepository interfaces.UsuarioRepository) interfaces.RegisterService {
	return &RegisterService{usuarioRepository: usuarioRepository}
}

// Register crea un nuevo usuario con rol productor y estado activo
func (s *RegisterService) Register(ctx context.Context, nombre, email, password, telefono string) (*entities.RegisterResponse, error) {
	exists, err := s.usuarioRepository.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	tel := telefono
	usuario := &entities.Usuario{
		Email:        email,
		PasswordHash: hashedPassword,
		Nombre:       nombre,
		Telefono:     &tel,
		Rol:          "productor",
		Estado:       "activo",
	}

	if err := s.usuarioRepository.Create(ctx, usuario); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &entities.RegisterResponse{
		IDUsuario:     usuario.ID,
		Nombre:        usuario.Nombre,
		Email:         usuario.Email,
		Rol:           usuario.Rol,
		FechaRegistro: usuario.FechaRegistro,
	}, nil
}
