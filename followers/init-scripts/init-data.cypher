// Create SocialProfile nodes
CREATE (a:SocialProfile {userId: -1, username: "admin"}),
       (a1:SocialProfile {userId: -11, username: "autor1"}),
       (a2:SocialProfile {userId: -12, username: "autor2"}),
       (a3:SocialProfile {userId: -13, username: "autor3"}),
       (t1:SocialProfile {userId: -21, username: "turista1"}),
       (t2:SocialProfile {userId: -22, username: "turista2"}),
       (t3:SocialProfile {userId: -23, username: "turista3"});

// Create FOLLOWS relationships
MATCH (a1:SocialProfile {userId: -11}), (a2:SocialProfile {userId: -12}) CREATE (a1)-[:FOLLOWS]->(a2)-[:FOLLOWS]->(a1);
MATCH (t1:SocialProfile {userId: -21}), (t2:SocialProfile {userId: -22}) CREATE (t1)-[:FOLLOWS]->(t2);
MATCH (t1:SocialProfile {userId: -21}), (t3:SocialProfile {userId: -23}) CREATE (t1)-[:FOLLOWS]->(t3)-[:FOLLOWS]->(t1);
MATCH (t2:SocialProfile {userId: -22}), (t3:SocialProfile {userId: -23}) CREATE (t2)-[:FOLLOWS]->(t3);
