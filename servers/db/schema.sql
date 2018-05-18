create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(255) not null,
    passhash binary(60) not null, 
    username varchar(255) not null, 
    firstname varchar(35) null,
    lastname varchar(35) null,
    photourl varchar(2083) not null,
    unique(email),       
    unique(username)   
);

create table if not exists userslogin (
    id int not null auto_increment primary key, 
    userid int not null,
    logintime datetime not null,
    ipaddr varchar(20) not null
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

create table if not exists channel_users (
    id int not null auto_increment primary key,
    channelid int not null,
    usersid int not null,
    foreign key(channelid) references channel(id),
    foreign key(usersid) references users(id),
    unique key (channelid, usersid)  
);

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

