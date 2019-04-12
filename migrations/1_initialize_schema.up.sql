CREATE TABLE public.users
(
    id uuid NOT NULL PRIMARY KEY,
    displayname text COLLATE pg_catalog."default",
    email text COLLATE pg_catalog."default",
    created date NOT NULL DEFAULT ('now'::text)::date,
    lastupdated date NOT NULL DEFAULT ('now'::text)::date
);

CREATE TABLE public.user_identities
(
    id uuid NOT NULL PRIMARY KEY,
    user_id text COLLATE pg_catalog."default" NOT NULL,
    source text COLLATE pg_catalog."default",
    externalid text COLLATE pg_catalog."default" NOT NULL,
    created date NOT NULL DEFAULT ('now'::text)::date,
    lastupdated date NOT NULL DEFAULT ('now'::text)::date
);

CREATE TABLE public.oauth_clients
(
    client_id uuid NOT NULL PRIMARY KEY,
    client_secret text COLLATE pg_catalog."default",
    redirect_uris text COLLATE pg_catalog."default",
    externalid text COLLATE pg_catalog."default" NOT NULL,
    is_private boolean NOT NULL,
    created date NOT NULL DEFAULT ('now'::text)::date,
    lastupdated date NOT NULL DEFAULT ('now'::text)::date
);