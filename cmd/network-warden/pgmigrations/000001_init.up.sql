begin;

create extension if not exists postgis;

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
  holder_id          bigint references holders (id) on delete cascade not null,
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

create table public.network_nodes
(
  id                            bigint primary key,
  created_at                    timestamp(0) with time zone default current_timestamp not null,
  last_modified_at              timestamp(0) with time zone default current_timestamp not null,
  network_warden_id             bigint not null,
  holder_id                     bigint references holders (id) on delete cascade not null,
  name                          text not null,
  description                   text not null,
  domain_name                   text not null,
  location                      geography(point, 4326),
  accounts_capacity             bigint not null,
  alive                         boolean not null,
  last_pinged_at                timestamp(0) with time zone default current_timestamp,
  is_open                       boolean not null,
  url                           text not null,
  api_key_hash                  text not null,
  version                       text not null,
  rate_limit_max_requests       bigint not null,
  rate_limit_interval           bigint not null,
  crawl_rate_limit_max_requests bigint not null,
  crawl_rate_limit_interval     bigint not null,
  status                        text not null,
  id_gen_node                   bigint not null
);
create unique index network_nodes_domain_name_uindex on network_nodes (domain_name);
create index network_nodes_status_index on network_nodes (status);

create table public.personal_data_nodes
(
  id                            bigint primary key,
  created_at                    timestamp(0) with time zone default current_timestamp not null,
  last_modified_at              timestamp(0) with time zone default current_timestamp not null,
  network_warden_id             bigint not null,
  holder_id                     bigint references holders (id) on delete cascade not null,
  label                         text not null,
  address                       text not null,
  name                          text not null,
  description                   text not null,
  location                      geography(point, 4326),
  accounts_capacity             bigint not null,
  alive                         boolean not null,
  last_pinged_at                timestamp(0) with time zone default current_timestamp,
  is_open                       boolean not null,
  url                           text not null,
  api_key_hash                  text not null,
  version                       text not null,
  rate_limit_max_requests       bigint not null,
  rate_limit_interval           bigint not null,
  crawl_rate_limit_max_requests bigint not null,
  crawl_rate_limit_interval     bigint not null,
  status                        text not null,
  id_gen_node                   bigint not null
);
create unique index personal_data_nodes_label_uindex on personal_data_nodes (label);
create unique index personal_data_nodes_address_uindex on personal_data_nodes (address);
create index personal_data_nodes_status_index on personal_data_nodes (status);

commit;
