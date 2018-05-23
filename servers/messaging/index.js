// "use strict";


//EXPORT MQADDR and MQNAME


const express = require("express");
const mysql = require("mysql");
const app = express();
const addr = process.env.ADDR || ":80";
const mqAddr = process.env.MQADDR;
const mqName = process.env.MQNAME;
const [host, port] = addr.split(":");
const portNum = parseInt(port);

var Channel = require("./models/channel");
var Message = require("./models/message");
var Constants = require("./constants");
var amqp = require('amqplib/callback_api');

var mqChannel;
var connection;
let numTries = 0;
let mqUrl = 'amqp://'+mqAddr;

let tryConn = setInterval(function(){
    connectToMQ(mqAddr);
}, 2000)

function connectToMQ(mqAddr){
    console.log("trying to dial amqp://"+mqAddr);
    amqp.connect('amqp://' + mqAddr, function(err, conn){
        if(err === null){
            console.log("successfully connected"); 
            conn.createChannel(function(err, ch) {
                ch.assertQueue(mqName, {durable: false});    
                mqChannel = ch;    
            });          
            clearInterval(tryConn);
        }else if(numTries === 20){
            console.log("unsucessful "+ err);
            clearInterval(tryConn);
        }else{
            console.log("tried " + numTries + " times");
            numTries++;
        }
    });
}




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

//query returns a simple promise that resolves to the rows
//returned by the database.
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
        let channelsDir = {};         
        let channels = [];
        db.query(Constants.SQL_GET_CHANNELS, (err, rows) => {
            if (err) {
                return next(err);
            }       
            rows.forEach((row) => {
                if (!channelsDir[row.channelid]){                                        
                    let newChannel = new Channel(row.channelid, row.channelname, row.channeldescription, row.channelprivate, row.createdat, 
                        {}, row.editedat);  
                    channelsDir[row.channelid] = newChannel;
                }
                if (row.channelprivate){   
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                      
                    channelsDir[row.channelid].pushMembers(user);
                }
                if (row.creatorid === row.usersid) {
                    let creator = {id: row.creatorid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};
                    channelsDir[row.channelid].setCreator(creator);
                }             

            });            
            channels = populateChannels(channelsDir);
            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
            return res.status(200).json(channels);      
        });            
    }
   

});

//populate channels only adds channels that the user is allowed to see
function populateChannels(channelsDir) {
    let channels = [];
    for (let key in channelsDir){
        let channel = channelsDir[key];                                
        if(channel.private && channel.containsUserID(authResult.id)) {
            channels.push(channel);                    
        }else if(!channel.private){
            channels.push(channel);                    
        }
    } 
    return channels;
}

//Handles POST /v1/channels
app.post("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {   
       
        if (req.body.name === undefined || req.body.name === "") {
            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
            return res.status(400).send("Provide a name for the channel");
        }                
        let newChannel = cleanRequest(req);
        newChannel.pushMembers(authResult);     

        //get members based on passed in members
        if (newChannel.private && req.body.members !== undefined){
            let map = getPostMembers(req);       
            db.query(Constants.SQL_POST_MEMBERS + map.params, map.ids, (err, rows) => {
                if (err) {
                    return next(err);
                }            
                if (rows.length === req.body.members.length) {
                    rows.forEach((row) => {
                        let user = {id: row.id, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                      
                        newChannel.pushMembers(user);
                    });                
                }else {
                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                    return res.status(400).send("User doesn't exist");
                }
            });
        }                        
        query(db, Constants.SQL_INSERT_CHANNEL, [newChannel.name, newChannel.description, newChannel.private, newChannel.createdAt, 
            authResult.id, newChannel.editedAt])
            .then(results => {                
                let newID = results.insertId;  
                newChannel.setId(newID); 
                return newID;               
            })
            .then((newID) => {                
                let membersMap = buildParams(req, newID, authResult, newChannel);
                newChannel = membersMap.newChannel;                              
                query(db, Constants.SQL_INSERT_MEMBER + membersMap.params, membersMap.ids)
                .catch((err) => {
                    if (err.message.startsWith(Constants.DUPLICATE_ENTRY)) {
                        res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                        return res.status(400).send("Member is already added");
                    }else if (err.message.startsWith(Constants.NO_REFERENCE)) {
                        res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                        return res.status(400).send("Could not find user");
                    }else{
                        return next(err);
                    }     
                });       
                
                mqChannelNotification("channel-new", newChannel);       

                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(201).json(newChannel);
            })
            .catch((err) =>{
                if (err.message.startsWith(Constants.DUPLICATE_ENTRY)) {
                    return res.status(400).send("Channel name is already added");
                }else{
                    return next(err);
                }    
            });
     }
});

//build ids and param to query users based on passed in userids
function getPostMembers(req){
    let map = {};   
    let members = req.body.members;
    let memberids = []
    memberids.push(req.body.members[0].id);
    let params = "(?"
    for (let i = 1; i < members.length; i++) {
        params += ", ?";
        memberids.push(members[i].id);
    }
    params += ")";   
    map.params = params;
    map.ids = memberids; 
    return map;         
    
}

//constructs new channel based on request
function cleanRequest(req){
    let description = "Temporary description";
    if (req.body.description !== undefined){
        description = req.body.description;
    }

    let setPrivate = false;
    if (req.body.private !== undefined) {
        setPrivate = req.body.private;
    }
    let time = getTimezoneTime();
    let newChannel = new Channel(0, req.body.name, description, setPrivate, time, 
        authResult, time); 
    return newChannel;
}

//builds params and user ids to insert members to associate entity
function buildParams(req, newID, authResult, newChannel){
    let membersMap = {};
    let members = [];
    let memberids = [];
    let params = " (?, ?)";
    memberids.push(newID);
    memberids.push(authResult.id);                 
    if (req.body.private && req.body.members != null){
        members = req.body.members;
        for(let i = 0; i < members.length; i++) {
            params += ",(?, ?)"
            memberids.push(newID);
            memberids.push(members[i].id);             
        }

    }else if (!req.body.private){ 
        newChannel.setMembers([]);
    }  
    membersMap.params = params;
    membersMap.ids = memberids;
    membersMap.newChannel = newChannel;
    return membersMap;
}

//Handles GET /v1/channels/{channelID}
app.get("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        try{            
            let inChannel = await checkUserInChannel(req, authResult);
            if (!inChannel) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Forbidden access to the channel");
            }            
            db.query(Constants.SQL_100_MESSAGES, [req.params.channelID], (err, rows) => {             
                if (err) {
                    return next(err);
                }                                      
                let messages = [];                
                rows.forEach((row) => {
                    let creator = {id: row.creatorid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                    
                    let message = new Message(row.id, row.channelid, row.body, row.createdat, creator, row.editedat);
                    messages.push(message);
                });
                
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(200).json(messages);          
            });  
        }catch(err) {
            if(err === "empty") {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(400).send("No information for this channel");
            }else{
                next(err);
            }            
        }                                  
    }
    
});



//Handles POST /v1/channels/{channelID}
app.post("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){                             
        try{                     
            //checks if channel has user
            let inChannel = await checkUserInChannel(req, authResult);
            if (!inChannel) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Forbidden access to the channel");
            }
                        
            if (req.body.body === undefined || req.body.body === ""){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Need to provide a body for the message");
            }
            let time = getTimezoneTime();
            //create a new message in this channel 
            let message = new Message(0, req.params.channelID, req.body.body, time, authResult, time);
            let channel = await getChannelMembers(req.params.channelID, authResult);
            db.query(Constants.SQL_INSERT_MESSAGE, [message.channelID, message.body, message.createdAt, message.creator.id, message.editedAt], (err, results) => {
                let newID = results.insertId;  
                message.setId(newID);
                                
                mqMessageNotification("message-new", message, channel.getUserIDs());

                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(201).json(message);  

            });    
        }catch(err) {
            if(err === "empty") {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(400).send("No information for this channel");
            }
            next(err);
        }

    }
});

function checkUserInChannel(req, authResult){
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                reject(err);
            }else if(rows.length == 0){
                reject("empty");
            }else {
                let result = rows[0];
                let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                result.createdat, {}, result.editedat);            
                rows.forEach((row) => {
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                    if(user.id === result.creatorid){
                        channel.setCreator(user);
                    }                  
                    channel.pushMembers(user);
                });
            
                if (channel.private && !channel.containsUserID(authResult.id)){
                    resolve(false);
                }
                resolve(true);            
            }                
            
        });
    });
}

//Handles PATCH /v1/channels/{channelID}
app.patch("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {        
        try{
            //checks if user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req.params.channelID, authResult);

            let newName = req.body.name;
            let newDesc = req.body.description;
            if ((newName == undefined || newName == "") && (newDesc == undefined || newDesc == "")){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Should update name and/or description of channel");
            }else if ((newName == undefined || newName == "") && (newDesc != undefined|| newDesc != "")) {
                newName = channel.name;
            }else if ((newName != undefined || newName != "" ) && (newDesc == undefined || newDesc == "")){
                newDesc = channel.description;
            }
            let time = getTimezoneTime();
            db.query(Constants.SQL_UPDATE_CHANNEL, [newName, newDesc, time, req.params.channelID], (err, results) => {
                if (err) {
                    return next(err);
                }

                db.query(Constants.SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }                    
                    channel.setName(rows[0].channelname);
                    channel.setDescription(rows[0].channeldescription);
                    channel.setEditedAt(rows[0].editedat);

                    mqChannelNotification("channel-update", channel);

                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                    res.status(200).json(channel);
                });                
            });               
        }catch(err) {
            if (err == "empty"){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(400).send("No information for this channel");
            }else{
                next(err);
            }            
        }
                        
    }    
});

//checks if the user is the creator of the channel
function checkIsCreator(req, authResult) {
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
            if(err) {
                reject(err);
            }else if (rows.length == 0) {
                reject("empty");
            }else {
                let result = rows[0];
                if (result.creatorid !== authResult.id) {
                    resolve(false);
                }  
                resolve(true);
            }
        });
    });
}

//gets members in the channel
function getChannelMembers(channelID, authResult) {
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_GET_MEMBERS, [channelID], (err, rows) => {
            if (err) {
                reject(err);
            }else if (rows.length == 0) {
                reject("empty");
            }else {
                let result = rows[0];         
                let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                    result.createdat, {}, result.editedat);
                channel.setCreator(authResult);
                if(result.channelprivate){
                    rows.forEach((row) => {
                        let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                        channel.pushMembers(user);
                    });                
                }
                resolve(channel);
            }
        });
    });
}

//notification for new channel and update channel
function mqChannelNotification(type, newChannel){
    let mqResult = {type: type, channel: newChannel, userIDs: newChannel.getUserIDs()};
    mqChannel.publish("", mqName, Buffer.from(JSON.stringify(mqResult)));
}

//Handles DELETE /v1/channels/{channelID}
app.delete("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        //checks if user is creator of channel
        try {            
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req.params.channelID, authResult);
            db.query(Constants.SQL_DELETE_CHANNEL_USERS, [req.params.channelID], (err, rows) => {
                if(err) {
                    return next(err);
                }

                db,query(Constants.SQL_DELETE_CHANNEL_MESSAGES, [req.params.channelID], (err, rows) => {
                    if(err) {
                        return next(err);
                    }
                    db.query(Constants.SQL_DELETE_CHANNEL, [req.params.channelID], (err, rows) => {
                        if(err) {
                            return next(err);
                        }                    
                    });
                });

                let mqResult = {type: "channel-delete", channelID: channel.getId(), userIDs: channel.getUserIDs()};
                mqChannel.publish('', mqName, Buffer.from(JSON.stringify(mqResult)));

                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(200).send("Channel deleted");
                
            });
        }catch(err) {
            if (err == "empty"){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(400).send("No information for this channel");
            }else{
                next(err);
            }
        }                               
                    
    }
});

//Handles POST /v1/channels/{channelID}/members
app.post("/v1/channels/:channelID/members", async (req, res, next) => {    
    authResult = checkAuthentication(req, res);
    if(authResult){                    
        try{
            //check user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req.params.channelID, authResult);
            if (req.body.id == undefined || req.body.id == null){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Should provide a user id to add to channel");
            }
            if(channel.private){
                let addUser = req.body;
                db.query(Constants.SQL_INSERT_MEMBER + "(?, ?)", [req.params.channelID, addUser.id], (err, rows) => {
                    if (err) {
                        if (err.message.startsWith(Constants.DUPLICATE_ENTRY)) {
                            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                            return res.status(400).send("Member is already added");
                        }else if (err.message.startsWith(Constants.NO_REFERENCE)) {
                            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                            return res.status(400).send("Could not find user");
                        }else{
                            return next(err);
                        }     
                    }
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(201).send("Member added to channel " + channel.name);  
                });   
            }else{
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Can't add user to public channel");
            }                                                                                                                                                                     
        }catch(err) {                     
            if (err == "empty"){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(400).send("No information for this channel");
            }else {
                next(err);
            }                                   
        }
    
    }
});

//Handles DELETE /v1/channels/{channelID}/members
app.delete("/v1/channels/:channelID/members", async (req, res, next) => {  
    authResult = checkAuthentication(req, res);
    if(authResult){                    
        try{            
            //check user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req.params.channelID, authResult);

            if (req.body.id == null){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Should provide a user id to add to channel");
            }

            if(channel.private){
                let addUser = req.body;
                db.query(Constants.SQL_DELETE_MEMBER, [req.params.channelID, addUser.id], (err, rows) => {
                if(err) {
                    return next(err);
                }
                if(rows.affectedRows === 0){
                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                    return res.status(400).send("User not in channel");
                }
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(200).send("Member deleted from channel " + channel.name);
                });  
            }else{
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Can't delete user from public channel");
            }                                  
        }catch(err) {
            if (err == "empty") {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("No information on channel");
            }else{
                next(err);
            }
            
        }             
    }
});


//Handles PATCH /v1/messages/{messageID}
app.patch("/v1/messages/:messageID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){             
        try{
            //check if user is creator of message
            let message = await checkMessageCreator(req, authResult);
            if(message === false) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to this message since you are not the creator");
            }

            if(req.body.body == undefined || req.body.body == ""){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("No message body provided");
            }
            let newBody = req.body.body;
            let time = getTimezoneTime();
            let channel = await getChannelMembers(message.getChannelID(), authResult);
            db.query(Constants.SQL_UPDATE_MESSAGE, [newBody, time, req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                db.query(Constants.SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }     
                    message.setBody(rows[0].body);
                    message.setEditedAt(rows[0].editedat);        
                    
                    mqMessageNotification("message-update", message, channel.getUserIDs());

                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                    res.status(200).json(message);
                });    
            });       
        }catch(err) {
            if(err === "empty"){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Cannot find message");
            }else{
                next(err);
            }            
        }
    }
});



//Handles DELETE /v1/messages/{messageID}
app.delete("/v1/messages/:messageID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){               
        try {
            //check if user is creator of message
            let message = await checkMessageCreator(req, authResult);
            if(message === false) {
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(403).send("Can't make changes to this message since you are not the creator");
            }
            let channel = await getChannelMembers(message.getChannelID(), authResult);
            //delete message. plain text message that it was successful            
            db.query(Constants.SQL_DELETE_MESSAGE, [req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                
                let mqResult = {type: "message-delete", messageID: message.getId(), userIDs: channel.getUserIDs()};
                mqChannel.publish('', mqName, Buffer.from(JSON.stringify(mqResult)));

                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(200).send("Message deleted");
            });                 
        }catch(err) {
            if(err === "empty"){
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                return res.status(400).send("Cannot find message");
            }else{
                return next(err);
            }  
        }        
    }             
});

//checks if the user is the message creator
function checkMessageCreator(req, authResult){
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
            if(err) {
                reject(err);
            }else if(rows.length == 0){
                reject("empty");
            }else{            
                let result = rows[0];
                let message = new Message(result.id, result.channelid, result.body, result.createdat, {}, result.editedat);
                if(result.creatorid !== authResult.id) {
                    resolve(false);
                }
                message.setCreator(authResult);
                resolve(message);
            }
        }); 
    })
}

//notification for new message and update message
function mqMessageNotification(type, newMessage, userIDs){
    let mqResult = {type: type, message: newMessage, userIDs: userIDs};
    mqChannel.publish("", mqName, Buffer.from(JSON.stringify(mqResult)));
}

app.use((err, req, res, next) => {
    if (err.stack) {
        console.error(err.stack);
    }
    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
    return res.status(500).send("Error in the server: " + err);
});

//checks for the "X-User " header
function checkAuthentication(req, res){
    let userJSON = req.get("X-User");
    if (userJSON) {
        let user = JSON.parse(userJSON);        
        return user;
    }else{
        res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
        res.status(401).send("Please sign in");
        return null;
    }  
}


//getCurrentTime returns time based on time zone
function getTimezoneTime() {    
    let starttime = new Date();
    let isotime = new Date((new Date(starttime)).toISOString() );
    let fixedtime = new Date(isotime.getTime()-(starttime.getTimezoneOffset()*60000));
    let currentTime = fixedtime.toISOString().slice(0, 19).replace('T', ' ');   
    return currentTime;
}