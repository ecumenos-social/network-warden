begin;

drop table if exists holder_sessions cascade;
drop table if exists holders cascade;
drop table if exists sent_emails cascade;

drop table if exists network_nodes cascade;
drop table if exists personal_data_nodes cascade;
drop table if exists network_wardens cascade;

drop table if exists admin_sessions;
drop table if exists admins;

commit;
