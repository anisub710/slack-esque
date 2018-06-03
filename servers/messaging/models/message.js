module.exports = class Message {
    constructor(id, channelID, body, createdAt, creator, editedAt) {
        this.id = id;
        this.channelID = channelID;
        this.body = body;        
        this.createdAt = createdAt;        
        this.creator = creator;                
        this.editedAt = editedAt;
        this.reactions = [];
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

    getReactions() {
        return this.reactions;
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

    setReactions(reactions) {
        this.reactions = reactions;
    }
    
    pushReaction(reaction){
        this.reactions = this.reactions.concat(reaction);  
    }
}
