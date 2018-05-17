// "use strict";

const express = require("express");
const mysql = require("mysql");
const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);

var Channel = require("./models/channel");
var Message = require("./models/message");


if (isNaN(portNum)) {
    throw new Error("port number is not a number");
}

let db = mysql.createPool({
    host: process.env.MYSQL_ADDR,
    database: process.env.MYSQL_DATABASE,
    user: "root",
    password: process.env.MYSQL_ROOT_PASSWORD
});

app.use(express.json());


const SQL_GET_CHANNELS = "select * from channel c " + 
                            "join channel_users cu on c.id = cu.channelid " + 
                            "join users u on u.id = cu.usersid " + 
                            "order by cu.channelid;";
const SQL_INSERT_CHANNEL = "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?);";

const SQL_INSERT_MEMBER = "insert into channel_users (channelid, usersid) values (?, ?);"  

const SQL_SELECT_CHANNEL = "select * from channel " +
                            "where id = ?;"

const SQL_100_MESSAGES = "select * from channel c " + 
                            "join channel_users cu on cu.channelid = c.id " + 
                            "join messages m on m.channelid = c.id " +  
                            "where c.id = ? " + 
                            "order by m.createdat " + 
                            "limit 100;";
const SQL_GET_MEMBERS = "select * from channel c " + 
                            "join channel_users cu on cu.channelid = c.id "+
                            "where c.id = ?;";
const SQL_INSERT_MESSAGE = "insert into messages (channelid, body, createdat, creatorid, editedat) "+ 
                            "values (?, ?, ?, ?, ?);"
const SQL_GET_MESSAGE = "select * from messages "+
                            "where id = ?;"
const SQL_UPDATE_CHANNEL = "update channel set channelname = ?, channeldescription = ? where id = ?;";

const CONTENT_TYPE = "Content-Type";
const CONTENT_JSON = "application/json";


function query(db, sql, params) {
    return new Promise((resolve, reject) => {
        db.query(sql, params, (err, results) => {
            if (err) {
                reject(err);
            }else {
                resolve(results);
            }
        });
    });
}


//start the server listening on host:port
app.listen(port, host, () => {
    //callback is executed once server is listening
    console.log(`server is listening at http://${addr}...`);
});

//Handles GET /v1/channels
app.get("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        
        db.query(SQL_GET_CHANNELS, (err, rows) => {
            if (err) {
                return next(err);
            }            
            let channels = [];
            let currChannel = rows[0].channelid;            
            let newChannel = new Channel(rows[0].channelid, rows[0].channelname, rows[0].channeldescription, rows[0].channelprivate, rows[0].createdat, 
                        rows[0].creatorid, rows[0].editedat);                
            let users = [];
            rows.forEach((row) => {  
                console.log(row);
                //TODO:check if user is allowed to see the channel.          
                if (currChannel !== row.channelid){
                    newChannel.setMembers(users);
                    users = [];
                    channels.push(newChannel);
                    currChannel = row.channelid
                    newChannel = new Channel(row.channelid, row.channelname, row.channeldescription, row.channelprivate, row.createdat, 
                                row.creatorid, row.editedat);                                                   
                }
                if (row.channelprivate){
                    users.push(row.usersid);                    
                }
    
            }); 
            newChannel.setMembers(users);
            channels.push(newChannel);            
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            res.status(200).json(channels);      
        });            
    }
   

});

//Handles POST /v1/channels
app.post("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);

    if (req.body.name === "") {
       return res.status(400).send("Provide a name for the channel");
    }

    let time = getTimezoneTime();
    if (authResult) {
        let members = [];
        let newChannel = new Channel(0, req.body.name, req.body.description, req.body.private, time, 
            authResult.id, time);   
          
        query(db, SQL_INSERT_CHANNEL, [req.body.name, req.body.description, req.body.private, time, 
            authResult.id, time])
            .then(results => {
                console.log(results);
                let newID = results.insertId;  
                newChannel.setId(newID); 
                return newID;               
            })
            .then((newID) => {
                console.log("second then");
                query(db, SQL_INSERT_MEMBER, [newID, authResult.id])
                .catch(next);
                members.push(authResult.id);
                newChannel.setMembers(members);                
                console.log(newChannel);
                res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                res.status(201).json(newChannel);
            })
            .catch(next);
        // db.query(SQL_INSERT_CHANNEL, [req.body.name, req.body.description, req.body.private, time, 
        //     authResult.id, time], (err, results) => {
        //         if (err) {
        //             return next(err);
        //         }
                
        //         let newID = results.insertId;  
        //         newChannel.setId(newID);                
                                                   
        //         db.query(SQL_INSERT_MEMBER, [newID, authResult.id], (err, results) => {
        //             if (err) {
        //                 return next(err);
        //             }                    
                                                    
        //         }); 
        //         members.push(authResult.id);
        //         newChannel.setMembers(members);                

        //     res.setHeader(CONTENT_TYPE, CONTENT_JSON);
        //     res.status(201).json(newChannel);   
        // });
    }
});

//Handles GET /v1/channels/{channelID}
app.get("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        db.query(SQL_100_MESSAGES, [req.params.channelID], (err, rows) => {
            if (err) {
                return next(err);
            }
            console.log(rows);
            let result = rows[0];
            if (result.private && !checkUserInChannel(rows, authResult.id)) {
                return res.status(403).send("Forbidden access to channel");
            }
            let messages = [];
            rows.forEach((row) => {
                let message = new Message(row.id, row.channelid, row.body, row.createdat, row.creatorid, row.editedat);
                messages.push(message);
            });
            
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            res.status(200).json(messages);  
        });
    }
});

//Handles POST /v1/channels/{channelID}
app.post("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult){
        //query to check if current user is not a member and respond with 403
        query(db, SQL_GET_MEMBERS, [req.params.channelID])
        .then(rows => {

            rows.forEach((row) => {
                if (row.usersid === authResult.id){
                    console.log("here");
                    return true;
                }
            });            
        })
        .then((inChannel) => {
            console.log("inChannel " + inChannel);
            let body = req.body.body;
            let time = getTimezoneTime();
            query(db, SQL_INSERT_MESSAGE, [req.params.channelID, body, time, authResult.id, time])
            .then((results) => {
                let newID = results.insertId;  
                console.log(results);                                               
                return newID;
            })
            .catch(next);
        })
        .catch(next);
        // db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
        //     if (err) {
        //         return next(err);
        //     }
        //     console.log(rows);
        //     // console.log(checkUserInChannel(rows, authResult.id))
        //     // if(!checkUserInChannel(rows, authResult.id)) {
        //     //     return res.status(403).send("Forbidden access to channel");
        //     // }
        //     rows.forEach((row) => {
        //         if (row.usersid === userID){
        //             console.log("here");
        //             let body = req.body.body;
        //             let time = getTimezoneTime();
        //             db.query(SQL_INSERT_MESSAGE, [req.params.channelID, body, time, authResult.id, time], (err, results) => {
        //                 if (err) {
        //                     return next(err);
        //                 }
                        
        //                 let newID = results.insertId;                                                 
        //                 db.query(SQL_GET_MESSAGE, [newID], (err, rows) => {
        //                     if (err) {
        //                         return next(err);
        //                     }
        //                     let result = rows[0];
        //                     let newMessage = new Message(result.id, result.channelid, result.body, result.createdat, 
        //                         result.creatorid, result.editedat);
                                
        //                     res.setHeader(CONTENT_TYPE, CONTENT_JSON);
        //                     res.status(201).json(newMessage); 
        //                 })
        //             });
        //         }
        //     });
        // });


    }
});

//Handles PATCH /v1/channels/{channelID}
app.patch("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        if (!checkCreator(req.params.channelID, authResult.id)) {
            return res.status(403).send("Can't make changes to channel since you are not the creator");
        }
        // TODO: check name AND/OR description
        db.query(SQL_UPDATE_CHANNEL, [req.body.name, req.body.description, req.params.channelid], (err, results) => {
            if (err) {
                return next(err);
            }
            //get last update id and select that channel, populate new channel and return
            // db.query(SQL_SELECT_CHANNEL, [])
        });
    }    
    
    // update only the name AND/OR description using the JSON in the request body and 
    // respond with a copy of the newly-updated channel
});

//Handles DELETE /v1/channels/{channelID}
app.delete("/v1/channels/:channelID", (req, res, next) => {

});

function checkUserInChannel(rows, userID){
    rows.forEach((row) => {
        if (row.usersid === userID){
            console.log("here");
            return true;
        }
    });
    return false;
}

app.use((err, req, res, next) => {
    if (err.stack) {
        console.error(err.stack);
    }
    res.status(500).send("Error in the server");
});

function checkAuthentication(req, res){
    let userJSON = req.get("X-User");
    if (userJSON) {
        let user = JSON.parse(userJSON);        
        return user;
    }else{
        res.status(401).send("Please sign in");
        return null;
    }  
}

function checkCreator(channelID, userID){
    //used in first get. refactor?
    db.query(SQL_SELECT_CHANNEL, [channelID], (err, rows) => {
        //TODO: check error
        if (rows[0].creatorid == userID) {
            return true;
        }        
    });
    return false;
}

//getCurrentTime returns time based on time zone
function getTimezoneTime() {    
    let starttime = new Date();
    let isotime = new Date((new Date(starttime)).toISOString() );
    let fixedtime = new Date(isotime.getTime()-(starttime.getTimezoneOffset()*60000));
    let currentTime = fixedtime.toISOString().slice(0, 19).replace('T', ' ');   
    return currentTime;
}