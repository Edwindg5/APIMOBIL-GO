package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
)

type ProfileService struct {
	usuarioRepository interfaces.UsuarioRepository
}

// NewProfileService crea una nueva instancia del servicio de perfil
func NewProfileService(usuarioRepository interfaces.UsuarioRepository) interfaces.ProfileService {
	return &ProfileService{usuarioRepository: usuarioRepository}
}

// GetProfile retorna el perfil del usuario autenticado sin exponer el password
func (s *ProfileService) GetProfile(ctx context.Context, userID int) (*entities.PerfilResponse, error) {
	usuario, err := s.usuarioRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if usuario == nil {
		return nil, errors.New("user not found")
	}

	return toPerfilResponse(usuario), nil
}

// UpdateProfile actualiza nombre y telefono; retorna el perfil actualizado
func (s *ProfileService) UpdateProfile(ctx context.Context, userID int, nombre, telefono string) (*entities.PerfilResponse, error) {
	usuario, err := s.usuarioRepository.Update(ctx, userID, nombre, telefono)
	if err != nil {
		return nil, fmt.Errorf("error updating profile: %w", err)
	}
	if usuario == nil {
		return nil, errors.New("user not found")
	}

	return toPerfilResponse(usuario), nil
}

// ChangePassword valida password_actual y establece password_nueva hasheada
func (s *ProfileService) ChangePassword(ctx context.Context, userID int, passwordActual, passwordNueva string) error {
	usuario, err := s.usuarioRepository.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if usuario == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(passwordActual)); err != nil {
		return errors.New("invalid current password")
	}

	hashedPassword, err := HashPassword(passwordNueva)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	if err := s.usuarioRepository.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	return nil
}

func toPerfilResponse(u *entities.Usuario) *entities.PerfilResponse {
	return &entities.PerfilResponse{
		IDUsuario:     u.ID,
		Nombre:        u.Nombre,
		Email:         u.Email,
		Rol:           u.Rol,
		Telefono:      u.Telefono,
		Estado:        u.Estado,
		FechaRegistro: u.FechaRegistro,
	}
}
