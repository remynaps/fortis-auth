
CREATE TABLE public.user_consent
(
    id uuid NOT NULL PRIMARY KEY,
    client_id uuid NOT NULL,
    user_id uuid NOT NULL ,
    created date NOT NULL DEFAULT ('now'::text)::date,
    lastupdated date NOT NULL DEFAULT ('now'::text)::date
);