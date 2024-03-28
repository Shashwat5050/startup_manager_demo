begin;
CREATE EXTENSION if not exists "uuid-ossp";
create table if not exists startups_info(
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    server_id uuid not null,
    variables JSONB,
    command   text,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT startups_servers_id_fk FOREIGN key(server_id) references gs_info(id) ON DELETE CASCADE
);
commit;