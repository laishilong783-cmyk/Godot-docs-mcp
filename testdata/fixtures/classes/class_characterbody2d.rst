CharacterBody2D
===============

**Inherits:** :ref:`PhysicsBody2D <class_PhysicsBody2D>`

A 2D physics body specialized for characters moved by script.

Description
-----------

CharacterBody2D is a specialized class for physics bodies that are meant to be user-controlled. They are not affected by physics at all, but they affect other physics bodies in their path.

Properties
----------

+----------+---------+-----------+
| Type     | Name    | Default   |
+==========+=========+===========+
| Vector2  | velocity| Vector2(0,0)|
+----------+---------+-----------+

Methods
-------

+-----------------------------------+------+
| Return type                       | Name |
+===================================+======+
| bool                              | move_and_slide |
+-----------------------------------+------+

Method Descriptions
-------------------

.. _class_CharacterBody2D_method_move_and_slide:

- bool **move_and_slide** ()

Moves the body based on velocity. Returns true if collided.
