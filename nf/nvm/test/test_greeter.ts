interface Person {
    firstName: string;
    lastName: string;
}

function greeter(person: Person) {
    return "Hello, " + person.firstName + " " + person.lastName;
}

let person = { firstName: "robin", lastName: "zhong" };

console.log(greeter(person));
