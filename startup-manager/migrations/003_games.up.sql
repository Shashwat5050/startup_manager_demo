begin;
CREATE EXTENSION if not exists "uuid-ossp";

create table games (
    id uuid not null default uuid_generate_v4(),
    name text not null,
    description text not null,
    image text not null,
    envs text[] not null default '{}',
    ports int[] not null default '{}',
    volumes text[] not null default ARRAY['/opt/'],
    cpu int not null default 500,
    memory int not null default 1024,
    command text not null default '',
    args text[] not null default '{}',
    default_startup_command text,
    default_variables  text[] not null,
    with_db boolean not null default false,
    created_at timestamp not null default now(),
    updated_at timestamp,
    primary key (id)
);

INSERT INTO games (name, description, image, envs, ports, volumes, cpu, memory, command, args,default_startup_command,default_variables,with_db, created_at, updated_at)
VALUES (
    'CS2 Server',
    'Counter-Strike 2 Server',
    'joedwards32/cs2',
    ARRAY['SRCDS_TOKEN=""', 'CS2_SERVERNAME="changeme"', 'CS2_IP=""', 'CS2_PORT=27015', 'CS2_RCON_PORT=""', 'CS2_LAN="0"', 'CS2_RCONPW="changeme"', 'CS2_MAXPLAYERS=10', 'CS2_ADDITIONAL_ARGS=""', 'CS2_GAMEMODE=1', 'CS2_STARTMAP="de_inferno"', 'CS2_LOG="on"'],
    '{}',
    ARRAY['/opt/'],
    500,
    1024,
    '',
    '{}',
    'srcds_run -game {{GAME_TYPE}} +map {{MAP}} +maxplayers {{MAX_PLAYERS}} -autoupdate',
    ARRAY['GAME_TYPE="0"','MAP="cs2-server"','MAX_PLAYERS="10"'],
    false,
    NOW(),
    NULL
);

commit;