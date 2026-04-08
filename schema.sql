CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS otp_codes (
    email VARCHAR(255) PRIMARY KEY,
    code VARCHAR(9) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE IF NOT EXISTS sync_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    client_profile_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    UNIQUE(user_id, client_profile_id)
);

CREATE TABLE IF NOT EXISTS sync_profile_settings (
    profile_id UUID PRIMARY KEY REFERENCES sync_profiles(id) ON DELETE CASCADE,
    default_url TEXT,
    animation_speed REAL,
    is_sharp_mode BOOLEAN,
    cursor_container_size REAL,
    cursor_pointer_size REAL,
    cursor_tracking_speed REAL,
    show_suggestions BOOLEAN,
    closed_tab_history_size REAL,
    back_square_offset_x REAL,
    back_square_offset_y REAL,
    back_square_idle_opacity REAL,
    search_engine INT,
    is_fullscreen_mode BOOLEAN,
    highlight_color INT,
    is_ad_block_enabled BOOLEAN,
    is_guide_mode_enabled BOOLEAN,
    is_desktop_mode BOOLEAN,
    is_enabled_media_control BOOLEAN,
    is_enabled_out_sync BOOLEAN,
    options_order TEXT,
    settings_order TEXT,
    hidden_options TEXT
);

CREATE TABLE IF NOT EXISTS sync_pinned_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES sync_profiles(id) ON DELETE CASCADE,
    client_app_id BIGINT NOT NULL,
    label VARCHAR(255),
    url TEXT,
    icon_url TEXT
);

CREATE TABLE IF NOT EXISTS sync_visited_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES sync_profiles(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT,
    UNIQUE(profile_id, url)
);
