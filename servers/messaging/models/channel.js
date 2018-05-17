module.exports =  class Channel {
    
    constructor(id, name, description, newPrivate, createdAt, creator, editedAt) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.private = newPrivate;
        this.members = [];
        this.createdAt = createdAt;        
        this.creator = creator;
        this.editedAt = editedAt;
    }

    getId() {
        return this.id;
    }
    getName() {
        return this.name;
    }
    getDescription() {
        return this.description;
    }
    getPrivate() {
        return this.private;
    }
    getMembers() {
        return this.members;
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
    setName(name) {
        this.name = name;
    }
    setDescription(description) {
        this.description = description
    }
    setPrivate(newPrivate) {
        this.private = newPrivate;
    }
    setMembers(members) {
        this.members = members;
    }
    pushMembers(member) {
        this.members = this.members.concat(member);  
      
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

    containsUserID(userID){
        let ids = this.members.map((member) => {
            return member.id
        });
        return ids.includes(userID)
    }


}
