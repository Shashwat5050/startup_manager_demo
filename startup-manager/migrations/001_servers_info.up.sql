begin;
CREATE EXTENSION if not exists "uuid-ossp";
create table if not exists gs_info (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    server_name text not null,
    game_name text not null,
    image text not null,
    command text not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
    
);


commit;
