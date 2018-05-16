var Message = class Message {
    constructor(id, channelID, body, createdAt, creator, editedAt) {
        this.id = id;
        this.channelID = channelID;
        this.body = body;        
        this.createdAt = createdAt;        
        this.creator = creator;
        this.editedAt = editedAt;
    }

    getId() {
        return this.id;
    }
    getChannelID() {
        return this.channelID;
    }
    getBody() {
        return this.body;
    }
    getCreatedAt() {
        return this.createdAt;
    }
    getCreator() {
        return this.creator;
    }
    getEditedAt() {
        return this.editedAt;        
    }

    setId(id) {
        this.id = id;
    }
    setChannelID(name) {
        this.channelID = channelID;
    }
    setBody(body) {
        this.body = body;
    }
    setCreatedAt(createdAt) {
        this.createdAt = createdAt;
    }
    setCreator(creator) {
        this.creator = creator;
    }
    setEditedAt(editedAt) {
        this.editedAt = editedAt;      
    }
}

module.exports.Message;