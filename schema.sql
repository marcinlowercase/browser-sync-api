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
    settings_json JSONB NOT NULL
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
