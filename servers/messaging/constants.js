module.exports = {
    SQL_GET_CHANNELS: "select * from channel c " + 
                        "join channel_users cu on c.id = cu.channelid " + 
                        "join users u on u.id = cu.usersid " + 
                        "order by cu.channelid;",
    SQL_INSERT_CHANNEL: "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?);",
    SQL_POST_MEMBERS: "select * from users where id in ",
    SQL_INSERT_MEMBER: "insert into channel_users (channelid, usersid) values ", 
    SQL_SELECT_CHANNEL: "select * from channel where id = ?",
    SQL_DELETE_CHANNEL:  "delete from channel where id = ?",
    SQL_DELETE_CHANNEL_USERS: "delete from channel_users where channelid = ?",
    SQL_DELETE_CHANNEL_MESSAGES: "delete from messages where channelid = ?",
    SQL_100_MESSAGES: "select * from users u " +                             
                        "join messages m on m.creatorid = u.id " +
                        "where m.channelid = ? " + 
                        "order by m.createdat " + 
                        "limit 100;",
    SQL_GET_MEMBERS: "select * from channel_users cu " + 
                        "join users u on u.id = cu.usersid " + 
                        "join channel c on c.id = cu.channelid "+                        
                        "where c.id = ?;",
    SQL_INSERT_MESSAGE: "insert into messages (channelid, body, createdat, creatorid, editedat) "+ 
                            "values (?, ?, ?, ?, ?);",
    SQL_GET_MESSAGE_WITH_CREATOR: "select * from messages m "+
                        "join users u on m.creatorid = u.id " + 
                        "where m.id = ?;",
    
    SQL_UPDATE_CHANNEL: "update channel set channelname = ?, channeldescription = ?, editedat = ? where id = ?;",
    SQL_DELETE_MEMBER: "delete from channel_users where channelid = ? and usersid = ?;",
    SQL_UPDATE_MESSAGE: "update messages set body = ?, editedat = ? where id = ?;",
    SQL_DELETE_MESSAGE: "delete from messages where id = ?;",
    SQL_DELETE_MESSAGE: "delete from messages where id = ?;",
    SQL_INSERT_REACTIONS: "insert into messages_reactions (messageid, userid, reaction) values (?, ?, ?);",
    SQL_GET_MESSAGE: "select * from messages m " +                                                      
                        "where id = ?;",
    SQL_GET_MESSAGE_WITH_REACTIONS: "select * from messages_reactions mr " +                                                                     
                                    "join users u on u.id = mr.userid " +
                                    "where mr.messageid = ?;",
    SQL_INSERT_STAR: "insert into starred_messages (userid, messageid) values (?, ?);",
    SQL_GET_STAR: "select * from starred_messages sm " +                    
                    "join messages m on m.id = sm.messagesid " + 
                    "join users u on u.id = m.creatorid " +
                    "where sm.userid = ?",
    SQL_DELETE_STAR: "delete from starred_messages where userid = ? and messageid = ?;",
    CONTENT_TYPE: "Content-Type",
    CONTENT_JSON: "application/json",
    CONTENT_TEXT: "text/plain",
    DUPLICATE_ENTRY: "ER_DUP_ENTRY",
    NO_REFERENCE: "ER_NO_REFERENCED_ROW"
}














