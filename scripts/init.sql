-- Crear tablas del sistema

-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nombre_completo VARCHAR(255),
    rol VARCHAR(50) NOT NULL DEFAULT 'productor',
    estado VARCHAR(50) NOT NULL DEFAULT 'activo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de sensores (ESP32)
CREATE TABLE IF NOT EXISTS sensores (
    id SERIAL PRIMARY KEY,
    esp32_id VARCHAR(255) UNIQUE NOT NULL,
    lote_id INTEGER,
    linked_at TIMESTAMP,
    last_seen TIMESTAMP,
    estado VARCHAR(50) NOT NULL DEFAULT 'inactivo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de lotes de café
CREATE TABLE IF NOT EXISTS lotes_cafe (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    nombre VARCHAR(255) NOT NULL,
    descripcion TEXT,
    area DECIMAL(10, 2),
    sensor_id INTEGER REFERENCES sensores(id),
    estado VARCHAR(50) NOT NULL DEFAULT 'activo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Agregar relación de lote_id a sensores después de crear lotes_cafe
ALTER TABLE sensores ADD CONSTRAINT fk_sensores_lote 
    FOREIGN KEY (lote_id) REFERENCES lotes_cafe(id) ON DELETE SET NULL;

-- Tabla de lecturas ambientales
CREATE TABLE IF NOT EXISTS lecturas_ambientales (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    sensor_id INTEGER NOT NULL REFERENCES sensores(id) ON DELETE CASCADE,
    temperatura DECIMAL(5, 2),
    humedad DECIMAL(5, 2),
    presion DECIMAL(8, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de alertas
CREATE TABLE IF NOT EXISTS alertas (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    tipo VARCHAR(50) NOT NULL,
    mensaje TEXT NOT NULL,
    nivel VARCHAR(50) NOT NULL DEFAULT 'info',
    leida BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de predicciones
CREATE TABLE IF NOT EXISTS predicciones (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    prediccion TEXT NOT NULL,
    probabilidad DECIMAL(5, 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de recomendaciones
CREATE TABLE IF NOT EXISTS recomendaciones (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    accion VARCHAR(255) NOT NULL,
    razon TEXT,
    prioridad VARCHAR(50) NOT NULL DEFAULT 'media',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de historial de eventos
CREATE TABLE IF NOT EXISTS historial_eventos (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    tipo VARCHAR(50) NOT NULL,
    descripcion TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de reportes
CREATE TABLE IF NOT EXISTS reportes (
    id SERIAL PRIMARY KEY,
    lote_id INTEGER NOT NULL REFERENCES lotes_cafe(id) ON DELETE CASCADE,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    tipo VARCHAR(50) NOT NULL,
    estado VARCHAR(50) NOT NULL DEFAULT 'pendiente',
    url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de audit log
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER REFERENCES usuarios(id) ON DELETE SET NULL,
    accion VARCHAR(255) NOT NULL,
    tabla VARCHAR(100),
    registro_id INTEGER,
    datos_anteriores JSONB,
    datos_nuevos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de provisioning tokens (para vincular dispositivos)
CREATE TABLE IF NOT EXISTS provisioning_tokens (
    id SERIAL PRIMARY KEY,
    esp32_id VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    usuario_id INTEGER NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    used_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para optimización
CREATE INDEX idx_usuarios_email ON usuarios(email);
CREATE INDEX idx_usuarios_rol ON usuarios(rol);
CREATE INDEX idx_lotes_cafe_usuario ON lotes_cafe(usuario_id);
CREATE INDEX idx_lotes_cafe_sensor ON lotes_cafe(sensor_id);
CREATE INDEX idx_sensores_esp32 ON sensores(esp32_id);
CREATE INDEX idx_sensores_lote ON sensores(lote_id);
CREATE INDEX idx_lecturas_lote ON lecturas_ambientales(lote_id);
CREATE INDEX idx_lecturas_sensor ON lecturas_ambientales(sensor_id);
CREATE INDEX idx_alertas_lote ON alertas(lote_id);
CREATE INDEX idx_alertas_leida ON alertas(leida);
CREATE INDEX idx_predicciones_lote ON predicciones(lote_id);
CREATE INDEX idx_recomendaciones_lote ON recomendaciones(lote_id);
CREATE INDEX idx_historial_lote ON historial_eventos(lote_id);
CREATE INDEX idx_reportes_usuario ON reportes(usuario_id);
CREATE INDEX idx_reportes_lote ON reportes(lote_id);
CREATE INDEX idx_audit_usuario ON audit_log(usuario_id);
CREATE INDEX idx_provisioning_tokens_usuario ON provisioning_tokens(usuario_id);
CREATE INDEX idx_provisioning_tokens_esp32 ON provisioning_tokens(esp32_id);

-- Crear extensión para UUID si es necesario
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- RLS: Habilitar Row Level Security en PostgreSQL (ejemplo para lotes_cafe)
-- Esto se configura dependiendo de tu estrategia de RLS
ALTER TABLE lotes_cafe ENABLE ROW LEVEL SECURITY;

-- Policy de RLS: Usuarios solo pueden ver sus propios lotes
CREATE POLICY lote_select_policy ON lotes_cafe FOR SELECT
    USING (usuario_id = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_insert_policy ON lotes_cafe FOR INSERT
    WITH CHECK (usuario_id = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_update_policy ON lotes_cafe FOR UPDATE
    USING (usuario_id = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_delete_policy ON lotes_cafe FOR DELETE
    USING (usuario_id = CAST(current_setting('app.current_user_id') AS INT));
