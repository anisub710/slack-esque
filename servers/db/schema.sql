create table if not exists users (
    id int not null auto_increment primary key CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    email varchar(255) not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,  
    passhash binary(60) not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    username varchar(255) not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    firstname varchar(35) null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,    
    lastname varchar(35) null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    photourl varchar(2083) not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    unique(email),       
    unique(username)   
);

create table if not exists userslogin (
    id int not null auto_increment primary key CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    userid int not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    logintime datetime not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    ipaddr varchar(20) not null CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
);

create table if not exists channel (
    id int not null auto_increment primary key,
    channelname varchar(255) not null,
    channeldescription varchar(255) not null,
    channelprivate boolean not null,
    createdat datetime not null,
    creatorid int not null,
    editedat datetime null,
    unique(channelname),
    foreign key(creatorid) references users(id)
);

create table if not exists members (
    id int not null auto_increment primary key,
    userid int not null 
);

create table if not exists channel__members (
    id int not null auto_increment primary key,
    channelid int not null,
    membersid int not null,
    foreign key(channelid) references channel(id),
    foreign key(membersid) references members(id)     
);

create table if not exists messages (
    id int not null auto_increment primary key,
    channelid int not null,
    body varchar(255) not null,
    createdat datetime not null,
    creatorid int not null,
    editedat datetime null,
    foreign key(channelid) references channel(id),
    foreign key(creatorid) references users(id)  
);

