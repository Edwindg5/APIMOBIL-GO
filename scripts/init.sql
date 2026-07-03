-- Esquema real de la base de datos en Neon (PostgreSQL).
-- Reescrito para coincidir exactamente con las columnas usadas por los repositorios en
-- internal/infrastructure/db/*.go. El script anterior (id, password, esp32_id, lote_id en
-- sensores, audit_log, provisioning_tokens, etc.) estaba obsoleto y no reflejaba la BD real.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- gen_random_uuid()

-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS usuarios (
    id_usuario     SERIAL PRIMARY KEY,
    nombre         VARCHAR(150) NOT NULL,
    email          VARCHAR(150) UNIQUE NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    rol            VARCHAR(20) NOT NULL DEFAULT 'productor'
                       CHECK (rol IN ('administrador', 'productor')),
    telefono       VARCHAR(20),
    estado         VARCHAR(20) NOT NULL DEFAULT 'activo'
                       CHECK (estado IN ('activo', 'inactivo')),
    fecha_registro TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de sensores (ESP32)
CREATE TABLE IF NOT EXISTS sensores (
    id_sensor          SERIAL PRIMARY KEY,
    mac_address        VARCHAR(50) UNIQUE NOT NULL,
    id_cola_mqtt       VARCHAR(150) NOT NULL,
    provisioning_token VARCHAR(255) UNIQUE,
    token_usado        BOOLEAN NOT NULL DEFAULT FALSE,
    tipo               VARCHAR(20) NOT NULL
                           CHECK (tipo IN ('temperatura', 'humedad', 'ambos')),
    modelo             VARCHAR(100),
    estado             VARCHAR(20) NOT NULL DEFAULT 'inactivo'
                           CHECK (estado IN ('activo', 'inactivo', 'mantenimiento')),
    fecha_registro     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ultima_conexion    TIMESTAMP
);

-- Tabla de lotes de café
CREATE TABLE IF NOT EXISTS lotes_cafe (
    id_lote             SERIAL PRIMARY KEY,
    id_usuario          INTEGER NOT NULL REFERENCES usuarios(id_usuario) ON DELETE CASCADE,
    id_sensor           INTEGER REFERENCES sensores(id_sensor),
    nombre_lote         VARCHAR(100) NOT NULL,
    variedad            VARCHAR(100),
    peso_kg             NUMERIC(10, 2),
    ubicacion           VARCHAR(200),
    codigo_qr           VARCHAR(100) UNIQUE NOT NULL,
    estado              VARCHAR(20) NOT NULL DEFAULT 'en_proceso'
                            CHECK (estado IN ('en_proceso', 'finalizado', 'cancelado')),
    fecha_inicio_secado TIMESTAMP,
    fecha_fin_secado    TIMESTAMP,
    linked_at           TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tipo_proceso        VARCHAR(50)
);

-- Tabla de lecturas ambientales
CREATE TABLE IF NOT EXISTS lecturas_ambientales (
    id_lectura  BIGSERIAL PRIMARY KEY,
    id_sensor   INTEGER NOT NULL REFERENCES sensores(id_sensor) ON DELETE CASCADE,
    id_lote     INTEGER NOT NULL REFERENCES lotes_cafe(id_lote) ON DELETE CASCADE,
    temperatura NUMERIC(5, 2),
    humedad     NUMERIC(5, 2),
    timestamp   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de alertas
CREATE TABLE IF NOT EXISTS alertas (
    id_alerta       SERIAL PRIMARY KEY,
    id_lote         INTEGER NOT NULL REFERENCES lotes_cafe(id_lote) ON DELETE CASCADE,
    id_sensor       INTEGER REFERENCES sensores(id_sensor),
    tipo_alerta     VARCHAR(100) NOT NULL,
    mensaje         TEXT,
    nivel_severidad VARCHAR(20) NOT NULL
                        CHECK (nivel_severidad IN ('baja', 'media', 'alta', 'critica')),
    atendida        BOOLEAN NOT NULL DEFAULT FALSE,
    fecha_generada  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_atencion  TIMESTAMP
);

-- Tabla de predicciones
-- id_modelo referencia una tabla "modelos" fuera del alcance de este esquema; se deja sin FK.
CREATE TABLE IF NOT EXISTS predicciones (
    id_prediccion         SERIAL PRIMARY KEY,
    id_lote               INTEGER NOT NULL REFERENCES lotes_cafe(id_lote) ON DELETE CASCADE,
    id_modelo             INTEGER NOT NULL,
    tiempo_estimado_horas NUMERIC(5, 2),
    calidad_estimada      VARCHAR(50),
    confianza             NUMERIC(5, 2),
    fecha_prediccion      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de recomendaciones
CREATE TABLE IF NOT EXISTS recomendaciones (
    id_recomendacion SERIAL PRIMARY KEY,
    id_lote          INTEGER NOT NULL REFERENCES lotes_cafe(id_lote) ON DELETE CASCADE,
    texto            TEXT NOT NULL,
    origen           VARCHAR(50) NOT NULL DEFAULT 'modelo_nlp',
    fecha_generada   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de historial de eventos
CREATE TABLE IF NOT EXISTS historial_eventos (
    id_evento    BIGSERIAL PRIMARY KEY,
    id_lote      INTEGER NOT NULL REFERENCES lotes_cafe(id_lote) ON DELETE CASCADE,
    id_usuario   INTEGER REFERENCES usuarios(id_usuario),
    tipo_evento  VARCHAR(100) NOT NULL,
    descripcion  TEXT,
    fecha_evento TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de reportes
CREATE TABLE IF NOT EXISTS reportes (
    id_reporte       SERIAL PRIMARY KEY,
    id_lote          INTEGER REFERENCES lotes_cafe(id_lote),
    id_usuario       INTEGER NOT NULL REFERENCES usuarios(id_usuario) ON DELETE CASCADE,
    tipo_reporte     VARCHAR(100),
    formato          VARCHAR(10) NOT NULL CHECK (formato IN ('pdf', 'excel')),
    url_archivo      VARCHAR(255),
    fecha_generacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Índices para optimización
CREATE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);
CREATE INDEX IF NOT EXISTS idx_usuarios_rol ON usuarios(rol);
CREATE INDEX IF NOT EXISTS idx_sensores_mac_address ON sensores(mac_address);
CREATE INDEX IF NOT EXISTS idx_lotes_cafe_usuario ON lotes_cafe(id_usuario);
CREATE INDEX IF NOT EXISTS idx_lotes_cafe_sensor ON lotes_cafe(id_sensor);
CREATE INDEX IF NOT EXISTS idx_lecturas_lote ON lecturas_ambientales(id_lote);
CREATE INDEX IF NOT EXISTS idx_lecturas_sensor ON lecturas_ambientales(id_sensor);
CREATE INDEX IF NOT EXISTS idx_alertas_lote ON alertas(id_lote);
CREATE INDEX IF NOT EXISTS idx_alertas_atendida ON alertas(atendida);
CREATE INDEX IF NOT EXISTS idx_predicciones_lote ON predicciones(id_lote);
CREATE INDEX IF NOT EXISTS idx_recomendaciones_lote ON recomendaciones(id_lote);
CREATE INDEX IF NOT EXISTS idx_historial_lote ON historial_eventos(id_lote);
CREATE INDEX IF NOT EXISTS idx_reportes_usuario ON reportes(id_usuario);
CREATE INDEX IF NOT EXISTS idx_reportes_lote ON reportes(id_lote);

-- RLS: los repositorios usan SET app.current_user_id por transacción (ver postgres.go)
ALTER TABLE lotes_cafe ENABLE ROW LEVEL SECURITY;

CREATE POLICY lote_select_policy ON lotes_cafe FOR SELECT
    USING (id_usuario = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_insert_policy ON lotes_cafe FOR INSERT
    WITH CHECK (id_usuario = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_update_policy ON lotes_cafe FOR UPDATE
    USING (id_usuario = CAST(current_setting('app.current_user_id') AS INT));

CREATE POLICY lote_delete_policy ON lotes_cafe FOR DELETE
    USING (id_usuario = CAST(current_setting('app.current_user_id') AS INT));
