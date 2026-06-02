class Base:
    @classmethod
    def who(cls):
        return cls.__name__

class Child(Base):
    pass

c = Child.create(5)
print(c.__class__.__name__)
