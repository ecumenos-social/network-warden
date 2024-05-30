begin;

create table public.holders
(
  id                bigint primary key,
  created_at        timestamp(0) with time zone default current_timestamp not null,
  last_modified_at  timestamp(0) with time zone default current_timestamp not null,
  emails            text array,
  phone_numbers     text array ,
  avatar_image_url  text,
  countries         text array,
  languages         text array,
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

create table public.sent_emails
(
  id               bigint primary key,
  created_at       timestamp(0) with time zone default current_timestamp not null,
  last_modified_at timestamp(0) with time zone default current_timestamp not null,
  sender_email     text not null,
  receiver_email   text not null,
  template_name    text not null
);
create index sent_emails_receiver_email_index on sent_emails (receiver_email);
create index sent_emails_template_name_index on sent_emails (template_name);

commit;
