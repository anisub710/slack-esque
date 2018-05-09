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

