create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(255) not null,
    passhash binary(60) not null, 
    username varchar(255) not null, 
    firstname varchar(35) null,
    lastname varchar(35) null,
    photourl varchar(2083) null,
    unique(email),       
    unique(username)   
);

create table if not exists userslogin (
    id int not null auto_increment primary key, 
    userid int not null,
    logintime datetime not null,
    ipaddr varchar(2083) not null
);

create table if not exists channel (
    id int not null auto_increment primary key,
    channelname varchar(255) not null,
    channeldescription varchar(255) null,
    channelprivate boolean not null,
    createdat datetime not null,
    creatorid int not null,
    editedat datetime null,
    unique(channelname),
    foreign key(creatorid) references users(id)
);

insert into users (id, email, passhash, username, firstname, lastname, photourl)
values(1, "system@email.com", "", "system", "", "", "");

insert into channel (id, channelname, channeldescription, channelprivate, createdat, creatorid, editedat)
values (1, "general", "channel for general things", false, LOCALTIME, 1, null);

create table if not exists channel_users (
    id int not null auto_increment primary key,
    channelid int not null,
    usersid int not null,
    foreign key(channelid) references channel(id),
    foreign key(usersid) references users(id),
    unique key (channelid, usersid)  
);

insert into channel_users (id, channelid, usersid) values (1, 1, 1);

create table if not exists messages (
    id int not null auto_increment primary key,
    channelid int not null,
    body varchar(4000) not null,
    createdat datetime not null,
    creatorid int not null,
    editedat datetime null,
    foreign key(channelid) references channel(id),
    foreign key(creatorid) references users(id)  
);

create table if not exists messages_reactions(
    id int not null auto_increment primary key,
    messageid int not null,
    userid int not null,
    reaction varchar(255) not null,
    foreign key(messageid) references messages(id),
    foreign key(userid) references users(id),
    unique key (messageid, userid, reaction)    
);

