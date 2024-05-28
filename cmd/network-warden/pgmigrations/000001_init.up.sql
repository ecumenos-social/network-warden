begin;

create table public.holders
(
  id                bigint primary key,
  created_at        timestamp(0) with time zone default current_timestamp not null,
  last_modified_at  timestamp(0) with time zone default current_timestamp not null,
  emails            text array not null,
  phone_numbers     text array not null,
  avatar_image_url  text,
  countries         text array not null,
  languages         text array not null,
  password_hash     text not null,
  confirmed         boolean not null,
  confirmation_code text not null
);
create unique index holders_emails_uindex on holders (emails);
create unique index holders_phone_numbers_uindex on holders (phone_numbers);

create table public.holder_sessions
(
  id                 bigint primary key,
  created_at         timestamp(0) with time zone default current_timestamp not null,
  last_modified_at   timestamp(0) with time zone default current_timestamp not null,
  holder_id          bigint references holders (id) not null,
  token              text not null,
  refresh_token      text not null,
  expired_at         timestamp(0) with time zone,
  remote_ip_address  text,
  remote_mac_address text
);

commit;
