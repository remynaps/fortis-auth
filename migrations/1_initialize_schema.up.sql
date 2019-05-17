CREATE TABLE public.users
(
    id uuid NOT NULL PRIMARY KEY,
    displayname text COLLATE pg_catalog."default",
    email text COLLATE pg_catalog."default",
    created date NOT NULL DEFAULT ('now'::text)::date,
    last_updated date NOT NULL DEFAULT ('now'::text)::date
);

CREATE TABLE public.user_identities
(
    id uuid NOT NULL PRIMARY KEY,
    user_id text COLLATE pg_catalog."default" NOT NULL,
    source text COLLATE pg_catalog."default",
    external_id text COLLATE pg_catalog."default" NOT NULL,
    created date NOT NULL DEFAULT ('now'::text)::date,
    last_updated date NOT NULL DEFAULT ('now'::text)::date
);

CREATE TABLE public.user_credentials
(
    id uuid NOT NULL PRIMARY KEY,
    user_id text COLLATE pg_catalog."default" NOT NULL,
    password text COLLATE pg_catalog."default" NOT NULL,
    compromised boolean COLLATE pg_catalog."default" NOT NULL,
    scheme_version int COLLATE pg_catalog."default" NOT NULL,
    last_updated date NOT NULL DEFAULT ('now'::text)::date
);

CREATE TABLE public.oauth_clients
(
    client_id uuid NOT NULL PRIMARY KEY,
    client_secret text COLLATE pg_catalog."default",
    display_name text COLLATE pg_catalog."default",
    redirect_uris text[] COLLATE pg_catalog."default",
    scopes text[] COLLATE pg_catalog."default",
    is_private boolean NOT NULL,
    created date NOT NULL DEFAULT ('now'::text)::date,
    last_updated date NOT NULL DEFAULT ('now'::text)::date
);