"use strict";

import {Channel} from "./models/channel"
import {Message} from "./models/message"

const express = require("express");
const mysql = require("mysql");
const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);


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


const SQL_GET_CHANNELS = "select * from channel_users cu " + 
                            "join channel c on c.id = cu.channelid " + 
                            "join users u on u.id = cu.usersid " + 
                            "order by cu.channelid;";
const SQL_INSERT_CHANNEL = "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?);";
const SQL_SELECT_CHANNEL = "select * from channel " +
                            "where id = ?;"

const SQL_100_MESSAGES = "select * from channel c " + 
                            "join channel_users cu on cu.channelid = c.id " + 
                            "join messages m on m.channelid = c.id " +  
                            "where c.id = ? " + 
                            "order by m.createdat " + 
                            "limit 100;";
const CONTENT_TYPE = "Content-Type";
const CONTENT_JSON = "application/json";

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
                if (currChannel !== row.channelid){
                    newChannel.setMembers(users);
                    users = [];
                    channels.push(newChannel);
                    currChannel = row.channelid
                    newChannel = new Channel(row.channelid, row.channelname, row.channeldescription, row.channelprivate, row.createdat, 
                                row.creatorid, row.editedat);                                                   
                }
                if (row.channelprivate){
                    users.push(row.userid);                    
                }
    
            }); 
            let channelsJSON = JSON.parse(channels);
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            res.status(200).json(channelsJSON);       
        });    
    }
   

});

//Handles POST /v1/channels
app.post("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);

    if (req.body.name === "") {
       return res.status(400).send("Provide a name for the channel");
    }

    let time = Date.now()
    if (authResult) {
        db.query(SQL_INSERT_CHANNEL, [req.body.name, req.body.description, req.body.private, time, 
            authResult.id, time], (err, results) => {
                if (err) {
                    return next(err);
                }
                
                let newID = results.insertId;                                                 
                //TODO: insert creator as member by adding user to channel_users   
                db.query(SQL_SELECT_CHANNEL, [newID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }
                    let result = rows[0];
                    let members = [];
                    let newChannel = new Channel(result.id, result.channelname, result.channeldescription, result.private, result.createdat, 
                        result.creatorid, result.editedat);
                    newChannel.setMembers(members.push(result.id))
                    let channelJSON = JSON.parse(newChannel);
                    res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                    res.status(201).json(channelJSON);  
                });                                                                     
        });
    }
});

//Handles GET /v1/channels/{channelID}
app.get("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        db.query(SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
            if (err) {
                return next(err);
            }
            let result = rows[0];
            if (result.private && !checkUserInChannel(rows, authResult.id)) {
                return res.status(403).send("Forbidden access to channel");
            }
            let messages = [];
            rows.forEach((row) => {
                let message = new Message(row.id, row.channelid, row.body, row.createdat, row.creatorid, row.editedat);
                messages.push(message);
            });
            let messageJSON = JSON.parse(messages)
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            res.status(200).json(messageJSON);  
        });
    }
});

//Handles POST /v1/channels/{channelID}
app.post("/v1/channels/:channelID", (req, res, next) => {
    //query to check if current user is not a member and respond with 403
    //only read body from request for message. set others based on context
    //respond with 201 and application.json. copy of new message midel with id
});

//Handles PATCH /v1/channels/{channelID}
app.patch("/v1/channels/:channelID", (req, res, next) => {

});

//Handles DELETE /v1/channels/{channelID}
app.delete("/v1/channels/:channelID", (req, res, next) => {

});

function checkUserInChannel(rows, userID){
    rows.forEach((row) => {
        if (row.userid === userID){
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
        res.json(user);
        return user;
    }else{
        res.status(401).send("Please sign in");
        return null;
    }  
}

// function getChannel(newID) {
//     let channel;
//     let members = [];
//     db.query(SQL_SELECT_CHANNEL, [newID], (err, rows) => {
//         if (err) {
//             return next(err);
//         }  
//        channel = new Channel(rows[0].channelid, rows[0].channelname, rows[0].channeldescription, rows[0].channelprivate, rows[0].createdat, 
//         rows[0].creatorid, rows[0].editedat);
//         rows.forEach((row) => {
//             members.push(row.userid)
//         })   
//         channel.setMembers(members)           
//     });    
//     return channel;
// }